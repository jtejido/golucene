package search

import (
	. "github.com/jtejido/golucene/core/search/model"
	"math"
)

type iDisjunctionScorer interface {
	/** Reset score state for a new match */
	reset()

	/** Factor in sub-scorer match */
	accum(subScorer Scorer) error

	/** Return final score */
	final() (float32, error)
}

/**
 * Base class for Scorers that score disjunctions.
 */
type DisjunctionScorer struct {
	*abstractScorer
	spi                   iDisjunctionScorer
	subScorers            []Scorer
	numScorers, doc, freq int
}

func newDisjunctionScorer(spi iDisjunctionScorer, weight Weight, subScorers []Scorer) (*DisjunctionScorer, error) {
	ans := &DisjunctionScorer{
		spi:        spi,
		doc:        -1,
		freq:       -1,
		subScorers: subScorers,
		numScorers: len(subScorers),
	}
	ans.abstractScorer = newScorer(ans, weight)

	ans.heapify()

	return ans, nil
}

func (s *DisjunctionScorer) heapify() {
	for i := (s.numScorers >> 1) - 1; i >= 0; i-- {
		s.heapAdjust(i)
	}
}

func (s *DisjunctionScorer) heapAdjust(root int) {
	scorer := s.subScorers[root]
	doc := scorer.DocId()
	i := root
	for i <= (s.numScorers>>1)-1 {
		lchild := (i << 1) + 1
		lscorer := s.subScorers[lchild]
		ldoc := lscorer.DocId()
		rdoc := int(math.MaxInt32)
		rchild := (i << 1) + 2
		var rscorer Scorer
		if rchild < s.numScorers {
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

func (s *DisjunctionScorer) heapRemoveRoot() {
	if s.numScorers == 1 {
		s.subScorers[0] = nil
		s.numScorers = 0
	} else {
		s.subScorers[0] = s.subScorers[s.numScorers-1]
		s.subScorers[s.numScorers-1] = nil
		s.numScorers--
		s.heapAdjust(0)
	}
}

// if we haven't already computed freq + score, do so
func (s *DisjunctionScorer) visitScorers() (err error) {
	s.spi.reset()
	s.freq = 1
	if err = s.spi.accum(s.subScorers[0]); err != nil {
		return
	}
	s.visit(1)
	s.visit(2)
	return
}

// TODO: remove recursion.
func (s *DisjunctionScorer) visit(root int) (err error) {
	if root < s.numScorers && s.subScorers[root].DocId() == s.doc {
		s.freq++
		if err = s.spi.accum(s.subScorers[root]); err != nil {
			return
		}
		s.visit((root << 1) + 1)
		s.visit((root << 1) + 2)
	}
	return
}

func (s *DisjunctionScorer) Score() (float32, error) {
	if err := s.visitScorers(); err != nil {
		return 0, err
	}
	return s.spi.final()
}

func (s *DisjunctionScorer) Freq() (n int, err error) {
	if s.freq < 0 {
		if err = s.visitScorers(); err != nil {
			return
		}
	}
	return s.freq, nil
}

func (s *DisjunctionScorer) DocId() int {
	return s.doc
}

func (s *DisjunctionScorer) NextDoc() (doc int, err error) {
	assert(s.doc != NO_MORE_DOCS)
	for {
		var d int
		d, err = s.subScorers[0].NextDoc()
		if err != nil {
			return
		}
		if d != NO_MORE_DOCS {
			s.heapAdjust(0)
		} else {
			s.heapRemoveRoot()
			if s.numScorers == 0 {
				s.doc = NO_MORE_DOCS
				return s.doc, nil
			}
		}
		docID := s.subScorers[0].DocId()
		if docID != s.doc {
			s.freq = -1
			s.doc = docID
			return s.doc, nil
		}
	}
}

func (s *DisjunctionScorer) Advance(target int) (doc int, err error) {
	assert(s.doc != NO_MORE_DOCS)
	for {
		var d int
		d, err = s.subScorers[0].Advance(target)
		if err != nil {
			return
		}
		if d != NO_MORE_DOCS {
			s.heapAdjust(0)
		} else {
			s.heapRemoveRoot()
			if s.numScorers == 0 {
				s.doc = NO_MORE_DOCS
				return s.doc, nil
			}
		}
		docID := s.subScorers[0].DocId()
		if docID >= target {
			s.freq = -1
			s.doc = docID
			return s.doc, nil
		}
	}
}

func (s *DisjunctionScorer) Cost() int64 {
	sum := int64(0)
	for i := 0; i < s.numScorers; i++ {
		sum += s.subScorers[i].Cost()
	}
	return sum
}
