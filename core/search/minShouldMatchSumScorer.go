package search

import (
	"fmt"
	. "github.com/jtejido/golucene/core/search/model"
	"github.com/jtejido/golucene/core/util"
	"math"
)

/**
 * A Scorer for OR like queries, counterpart of <code>ConjunctionScorer</code>.
 * This Scorer implements {@link Scorer#advance(int)} and uses advance() on the given Scorers.
 *
 * This implementation uses the minimumMatch constraint actively to efficiently
 * prune the number of candidates, it is hence a mixture between a pure DisjunctionScorer
 * and a ConjunctionScorer.
 */
type MinShouldMatchSumScorer struct {
	abstractScorer
	numScorers, mm, sortedSubScorersIdx, nrInHeap, doc, nrMatchers int
	sortedSubScorers, subScorers, mmStack                          []Scorer
	score                                                          float64
	coord                                                          []float32
}

type subScorersSorter []Scorer

func (s subScorersSorter) Len() int      { return len(s) }
func (s subScorersSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s subScorersSorter) Less(i, j int) bool {
	return math.Signbit(float64(s[j].Cost()) - float64(s[i].Cost()))
}

func newMinShouldMatchSumScorer(weight Weight, subScorers []Scorer, minimumNrMatchers int, coord []float32) (*MinShouldMatchSumScorer, error) {
	ans := &MinShouldMatchSumScorer{
		doc:                 -1,
		nrMatchers:          -1,
		score:               math.NaN(),
		nrInHeap:            len(subScorers),
		numScorers:          len(subScorers),
		mm:                  minimumNrMatchers,
		sortedSubScorers:    subScorers,
		mmStack:             make([]Scorer, minimumNrMatchers-1),
		sortedSubScorersIdx: minimumNrMatchers - 1,
		coord:               coord,
	}
	ans.weight = weight

	if minimumNrMatchers <= 0 {
		return nil, fmt.Errorf("Minimum nr of matchers must be positive")
	}
	if ans.numScorers <= 1 {
		return nil, fmt.Errorf("There must be at least 2 subScorers")
	}

	util.TimSort(subScorersSorter(ans.sortedSubScorers))
	for i := 0; i < ans.mm-1; i++ {
		ans.mmStack[i] = ans.sortedSubScorers[i]
	}
	ans.nrInHeap -= ans.mm - 1
	ans.subScorers = make([]Scorer, ans.nrInHeap)
	for i := 0; i < ans.nrInHeap; i++ {
		ans.subScorers[i] = ans.sortedSubScorers[ans.mm-1+i]
	}
	ans.minheapHeapify()
	assert(ans.minheapCheck())
	return ans, nil
}

func (s *MinShouldMatchSumScorer) minheapHeapify() {
	for i := (s.nrInHeap >> 1) - 1; i >= 0; i-- {
		s.minheapSiftDown(i)
	}
}

func (s *MinShouldMatchSumScorer) minheapSiftDown(root int) {
	// TODO could this implementation also move rather than swapping neighbours?
	scorer := s.subScorers[root]
	doc := scorer.DocId()
	i := root
	for i <= (s.nrInHeap>>1)-1 {
		lchild := (i << 1) + 1
		lscorer := s.subScorers[lchild]
		ldoc := lscorer.DocId()
		rdoc := int(math.MaxInt32)
		rchild := (i << 1) + 2
		var rscorer Scorer
		if rchild < s.nrInHeap {
			rscorer = s.subScorers[rchild]
			rdoc = rscorer.DocId()
		}
		if ldoc < doc {
			if rdoc < ldoc {
				s.subScorers[i] = rscorer
				s.subScorers[rchild] = scorer
				i = rchild
			} else {
				s.subScorers[i] = lscorer
				s.subScorers[lchild] = scorer
				i = lchild
			}
		} else if rdoc < doc {
			s.subScorers[i] = rscorer
			s.subScorers[rchild] = scorer
			i = rchild
		} else {
			return
		}
	}
}

func (s *MinShouldMatchSumScorer) minheapSiftUp(i int) {
	scorer := s.subScorers[i]
	doc := scorer.DocId()
	// find right place for scorer
	for i > 0 {
		parent := (i - 1) >> 1
		pscorer := s.subScorers[parent]
		pdoc := pscorer.DocId()
		if pdoc > doc { // move root down, make space
			s.subScorers[i] = s.subScorers[parent]
			i = parent
		} else { // done, found right place
			break
		}
	}
	s.subScorers[i] = scorer
}

func (s *MinShouldMatchSumScorer) minheapRemoveRoot() {
	if s.nrInHeap == 1 {
		//subScorers[0] = null; // not necessary
		s.nrInHeap = 0
	} else {
		s.nrInHeap--
		s.subScorers[0] = s.subScorers[s.nrInHeap]
		//subScorers[nrInHeap] = null; // not necessary
		s.minheapSiftDown(0)
	}
}

func (s *MinShouldMatchSumScorer) minheapRemove(scorer Scorer) bool {
	// find scorer: O(nrInHeap)
	for i := 0; i < s.nrInHeap; i++ {
		if s.subScorers[i] == scorer { // remove scorer
			s.nrInHeap--
			s.subScorers[i] = s.subScorers[s.nrInHeap]
			//if (i != nrInHeap) subScorers[nrInHeap] = null; // not necessary
			s.minheapSiftUp(i)
			s.minheapSiftDown(i)
			return true
		}
	}
	return false // scorer already exhausted
}

func (s *MinShouldMatchSumScorer) minheapCheck() bool {
	return s.minheapCheckFromX(0)
}

func (s *MinShouldMatchSumScorer) minheapCheckFromX(root int) bool {
	if root >= s.nrInHeap {
		return true
	}
	lchild := (root << 1) + 1
	rchild := (root << 1) + 2
	if lchild < s.nrInHeap && s.subScorers[root].DocId() > s.subScorers[lchild].DocId() {
		return false
	}
	if rchild < s.nrInHeap && s.subScorers[root].DocId() > s.subScorers[rchild].DocId() {
		return false
	}
	return s.minheapCheckFromX(lchild) && s.minheapCheckFromX(rchild)
}

func (s *MinShouldMatchSumScorer) NextDoc() (doc int, err error) {
	assert(s.doc != NO_MORE_DOCS)
	for {
		// to remove current doc, call next() on all subScorers on current doc within heap
		for s.subScorers[0].DocId() == s.doc {
			doc, err = s.subScorers[0].NextDoc()
			if err != nil {
				return 0, err
			}

			if doc != NO_MORE_DOCS {
				s.minheapSiftDown(0)
			} else {
				s.minheapRemoveRoot()
				s.numScorers--
				if s.numScorers < s.mm {
					s.doc = NO_MORE_DOCS
					return s.doc, nil
				}
			}
			//assert minheapCheck();
		}

		if err = s.evaluateSmallestDocInHeap(); err != nil {
			return
		}

		if s.nrMatchers >= s.mm { // doc satisfies mm constraint
			break
		}
	}
	return s.doc, nil
}

func (s *MinShouldMatchSumScorer) evaluateSmallestDocInHeap() error {
	// within heap, subScorer[0] now contains the next candidate doc
	var err error
	s.doc = s.subScorers[0].DocId()
	if s.doc == NO_MORE_DOCS {
		s.nrMatchers = int(math.MaxInt32) // stop looping
		return nil
	}
	// 1. score and count number of matching subScorers within heap
	var sc float32
	sc, err = s.subScorers[0].Score()
	s.score = float64(sc)
	if err != nil {
		return err
	}
	s.nrMatchers = 1
	s.countMatches(1)
	s.countMatches(2)
	// 2. score and count number of matching subScorers within stack,
	// short-circuit: stop when mm can't be reached for current doc, then perform on heap next()
	// TODO instead advance() might be possible, but complicates things
	for i := s.mm - 2; i >= 0; i-- { // first advance sparsest subScorer
		var d int
		d, err := s.mmStack[i].Advance(s.doc)
		if err != nil {
			return err
		}
		if s.mmStack[i].DocId() >= s.doc || d != NO_MORE_DOCS {
			if s.mmStack[i].DocId() == s.doc { // either it was already on doc, or got there via advance()
				s.nrMatchers++

				sc, err = s.mmStack[i].Score()
				if err != nil {
					return err
				}
				s.score += float64(sc)
			} else { // scorer advanced to next after doc, check if enough scorers left for current doc
				if s.nrMatchers+i < s.mm { // too few subScorers left, abort advancing
					return nil // continue looping TODO consider advance() here
				}
			}
		} else { // subScorer exhausted
			s.numScorers--
			if s.numScorers < s.mm { // too few subScorers left
				s.doc = NO_MORE_DOCS
				s.nrMatchers = int(math.MaxInt32) // stop looping
				return nil
			}
			if s.mm-2-i > 0 {
				// shift RHS of array left
				//System.arraycopy(mmStack, i+1, mmStack, i, mm-2-i)
				copy(s.mmStack[i:s.mm-2-i], s.mmStack[i+1:s.mm-2-i])
			}
			// find next most costly subScorer within heap TODO can this be done better?
			for !s.minheapRemove(s.sortedSubScorers[s.sortedSubScorersIdx]) {
				s.sortedSubScorersIdx++
				//assert minheapCheck();
			}
			// add the subScorer removed from heap to stack
			s.mmStack[s.mm-2] = s.sortedSubScorers[s.sortedSubScorersIdx-1]

			if s.nrMatchers+i < s.mm { // too few subScorers left, abort advancing
				return nil // continue looping TODO consider advance() here
			}
		}
	}

	return nil
}

func (s *MinShouldMatchSumScorer) countMatches(root int) error {
	if root < s.nrInHeap && s.subScorers[root].DocId() == s.doc {
		s.nrMatchers++
		sc, err := s.subScorers[root].Score()
		if err != nil {
			return err
		}
		s.score += float64(sc)
		s.countMatches((root << 1) + 1)
		s.countMatches((root << 1) + 2)
	}

	return nil
}

func (s *MinShouldMatchSumScorer) Score() (float32, error) {
	return s.coord[s.nrMatchers] * float32(s.score), nil
}

func (s *MinShouldMatchSumScorer) DocId() int {
	return s.doc
}

func (s *MinShouldMatchSumScorer) Freq() (n int, err error) {
	return s.nrMatchers, nil
}

func (s *MinShouldMatchSumScorer) Advance(target int) (doc int, err error) {
	if s.numScorers < s.mm {
		s.doc = NO_MORE_DOCS
		return s.doc, nil
	}
	// advance all Scorers in heap at smaller docs to at least target
	for s.subScorers[0].DocId() < target {
		doc, err = s.subScorers[0].Advance(target)
		if err != nil {
			return 0, err
		}
		if doc != NO_MORE_DOCS {
			s.minheapSiftDown(0)
		} else {
			s.minheapRemoveRoot()
			s.numScorers--
			if s.numScorers < s.mm {
				s.doc = NO_MORE_DOCS
				return s.doc, nil
			}
		}
		//assert minheapCheck();
	}

	s.evaluateSmallestDocInHeap()

	if s.nrMatchers >= s.mm {
		return s.doc, nil
	} else {
		return s.NextDoc()
	}
}

func (s *MinShouldMatchSumScorer) Cost() int64 {
	// cost for merging of lists analog to DisjunctionSumScorer
	costCandidateGeneration := int64(0)
	for i := 0; i < s.nrInHeap; i++ {
		costCandidateGeneration += s.subScorers[i].Cost()
	}
	// TODO is cost for advance() different to cost for iteration + heap merge
	//      and how do they compare overall to pure disjunctions?
	c1 := float32(1.0)
	c2 := float32(1.0) // maybe a constant, maybe a proportion between costCandidateGeneration and sum(subScorer_to_be_advanced.cost())?
	res := c1*float32(costCandidateGeneration) + c2*float32(costCandidateGeneration)*(float32(s.mm)-1)
	return int64(res)
}
