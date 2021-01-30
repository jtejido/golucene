package search

import (
	. "github.com/jtejido/golucene/core/search/model"
)

type ReqOptSumScorer struct {
	*abstractScorer
	reqScorer, optScorer Scorer
}

func newReqOptSumScorer(reqScorer, optScorer Scorer) (*ReqOptSumScorer, error) {
	assert(reqScorer != nil)
	assert(optScorer != nil)
	ans := &ReqOptSumScorer{
		reqScorer: reqScorer,
		optScorer: optScorer,
	}

	ans.abstractScorer = newScorer(ans, reqScorer.Weight())
	return ans, nil
}

func (s *ReqOptSumScorer) NextDoc() (doc int, err error) {
	return s.reqScorer.NextDoc()
}

func (s *ReqOptSumScorer) Advance(target int) (doc int, err error) {
	return s.reqScorer.Advance(target)
}

func (s *ReqOptSumScorer) DocId() int {
	return s.reqScorer.DocId()
}

/** Returns the score of the current document matching the query.
 * Initially invalid, until {@link #nextDoc()} is called the first time.
 * @return The score of the required scorer, eventually increased by the score
 * of the optional scorer when it also matches the current document.
 */
func (s *ReqOptSumScorer) Score() (reqScore float32, err error) {
	// TODO: sum into a double and cast to float if we ever send required clauses to BS1
	curDoc := s.reqScorer.DocId()
	reqScore, err = s.reqScorer.Score()
	if err != nil {
		return
	}
	if s.optScorer == nil {
		return
	}

	optScorerDoc := s.optScorer.DocId()
	optScorerDoc, err = s.optScorer.Advance(curDoc)
	if err != nil {
		return
	}
	if optScorerDoc < curDoc && optScorerDoc == NO_MORE_DOCS {
		s.optScorer = nil
		return
	}

	if optScorerDoc == curDoc {
		var sc float32
		sc, err = s.optScorer.Score()
		if err != nil {
			return
		}
		return reqScore + sc, nil
	}
	return
}

func (s *ReqOptSumScorer) Freq() (n int, err error) {
	// we might have deferred advance()
	s.Score()
	if s.optScorer != nil {
		if s.optScorer.DocId() == s.reqScorer.DocId() {
			return 2, nil
		}

	}
	return 1, nil
}

func (s *ReqOptSumScorer) Cost() int64 {
	return s.reqScorer.Cost()
}
