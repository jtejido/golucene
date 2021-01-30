package search

import (
	"github.com/jtejido/golucene/core/util"
)

/** Scorer for conjunctions, sets of queries, all of which are required. */
type ConjunctionScorer struct {
	*abstractScorer
	lastDoc      int
	docsAndFreqs []*DocsAndFreqs
	lead         *DocsAndFreqs
	coord        float32
}

type DocsAndFreqsSorter []*DocsAndFreqs

func (s DocsAndFreqsSorter) Len() int      { return len(s) }
func (s DocsAndFreqsSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s DocsAndFreqsSorter) Less(i, j int) bool {
	return s[i].cost < s[j].cost
}

func newConjunctionScorer(weight Weight, scorers []Scorer) (*ConjunctionScorer, error) {
	return newConjunctionScorerWithCoord(weight, scorers, 1)
}

func newConjunctionScorerWithCoord(weight Weight, scorers []Scorer, coord float32) (*ConjunctionScorer, error) {
	ans := &ConjunctionScorer{
		lastDoc:      -1,
		coord:        coord,
		docsAndFreqs: make([]*DocsAndFreqs, len(scorers)),
	}
	ans.abstractScorer = newScorer(ans, weight)

	for i := 0; i < len(scorers); i++ {
		ans.docsAndFreqs[i] = newDocsAndFreqs(scorers[i])
	}

	util.TimSort(DocsAndFreqsSorter(ans.docsAndFreqs))
	ans.lead = ans.docsAndFreqs[0]

	return ans, nil
}

func (s *ConjunctionScorer) doNext(doc int) (d int, err error) {
	for {
		// doc may already be NO_MORE_DOCS here, but we don't check explicitly
		// since all scorers should advance to NO_MORE_DOCS, match, then
		// return that value.
	advanceHead:
		for {
			for i := 1; i < len(s.docsAndFreqs); i++ {
				// invariant: docsAndFreqs[i].doc <= doc at this point.

				// docsAndFreqs[i].doc may already be equal to doc if we "broke advanceHead"
				// on the previous iteration and the advance on the lead scorer exactly matched.
				if s.docsAndFreqs[i].doc < doc {
					s.docsAndFreqs[i].doc, err = s.docsAndFreqs[i].scorer.Advance(doc)
					if err != nil {
						return 0, err
					}

					if s.docsAndFreqs[i].doc > doc {
						// DocsEnum beyond the current doc - break and advance lead to the new highest doc.
						doc = s.docsAndFreqs[i].doc
						break advanceHead
					}
				}
			}
			// success - all DocsEnums are on the same doc
			return doc, nil
		}
		// advance head for next iteration
		s.lead.doc, err = s.lead.scorer.Advance(doc)
		if err != nil {
			return 0, err
		}

		doc = s.lead.doc
	}
}

func (s *ConjunctionScorer) Advance(target int) (doc int, err error) {
	s.lead.doc, err = s.lead.scorer.Advance(target)
	if err != nil {
		return 0, err
	}

	s.lastDoc, err = s.doNext(s.lead.doc)
	return s.lastDoc, err
}

func (s *ConjunctionScorer) DocId() int {
	return s.lastDoc
}

func (s *ConjunctionScorer) NextDoc() (doc int, err error) {
	s.lead.doc, err = s.lead.scorer.NextDoc()
	if err != nil {
		return
	}
	s.lastDoc, err = s.doNext(s.lead.doc)
	return s.lastDoc, err
}

func (s *ConjunctionScorer) Score() (float32, error) {
	// TODO: sum into a double and cast to float if we ever send required clauses to BS1
	var sum float32
	for _, docs := range s.docsAndFreqs {
		sc, err := docs.scorer.Score()
		if err != nil {
			return 0, err
		}
		sum += sc
	}
	return sum * s.coord, nil
}

func (s *ConjunctionScorer) Freq() (n int, err error) {
	return len(s.docsAndFreqs), nil
}

func (s *ConjunctionScorer) Cost() int64 {
	return s.lead.scorer.Cost()
}

type DocsAndFreqs struct {
	cost   int64
	scorer Scorer
	doc    int
}

func newDocsAndFreqs(scorer Scorer) *DocsAndFreqs {
	return &DocsAndFreqs{
		scorer: scorer,
		cost:   scorer.Cost(),
	}
}
