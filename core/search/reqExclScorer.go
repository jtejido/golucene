package search

import (
	. "github.com/jtejido/golucene/core/search/model"
)

/** A Scorer for queries with a required subscorer
 * and an excluding (prohibited) sub DocIdSetIterator.
 * <br>
 * This <code>Scorer</code> implements {@link Scorer#advance(int)},
 * and it uses the skipTo() on the given scorers.
 */
type ReqExclScorer struct {
	abstractScorer
	reqScorer Scorer
	exclDisi  DocIdSetIterator
	doc       int
}

func newReqExclScorer(reqScorer Scorer, exclDisi DocIdSetIterator) (*ReqExclScorer, error) {
	ans := &ReqExclScorer{
		reqScorer: reqScorer,
		exclDisi:  exclDisi,
		doc:       -1,
	}

	ans.weight = reqScorer.Weight()
	return ans, nil
}

/** Advance to non excluded doc.
 * <br>On entry:
 * <ul>
 * <li>reqScorer != null,
 * <li>exclScorer != null,
 * <li>reqScorer was advanced once via next() or skipTo()
 *      and reqScorer.doc() may still be excluded.
 * </ul>
 * Advances reqScorer a non excluded required doc, if any.
 * @return true iff there is a non excluded required doc.
 */
func (s *ReqExclScorer) toNonExcluded() (doc int, err error) {
	exclDoc := s.exclDisi.DocId()
	reqDoc := s.reqScorer.DocId() // may be excluded
	for {
		if reqDoc < exclDoc {
			return reqDoc, nil // reqScorer advanced to before exclScorer, ie. not excluded
		} else if reqDoc > exclDoc {
			exclDoc, err = s.exclDisi.Advance(reqDoc)
			if err != nil {
				return
			}
			if exclDoc == NO_MORE_DOCS {
				s.exclDisi = nil // exhausted, no more exclusions
				return reqDoc, nil
			}
			if exclDoc > reqDoc {
				return reqDoc, nil // not excluded
			}
		}
		reqDoc, err = s.reqScorer.NextDoc()
		if err != nil {
			return
		}
		if reqDoc == NO_MORE_DOCS {
			break
		}
	}
	s.reqScorer = nil // exhausted, nothing left
	return NO_MORE_DOCS, nil
}

func (s *ReqExclScorer) NextDoc() (doc int, err error) {
	if s.reqScorer == nil {
		return s.doc, nil
	}
	s.doc, err = s.reqScorer.NextDoc()
	if err != nil {
		return
	}
	if s.doc == NO_MORE_DOCS {
		s.reqScorer = nil // exhausted, nothing left
		return s.doc, nil
	}
	if s.exclDisi == nil {
		return s.doc, nil
	}
	s.doc, err = s.toNonExcluded()
	return s.doc, err
}

func (s *ReqExclScorer) Advance(target int) (doc int, err error) {
	var d int
	if s.reqScorer == nil {
		s.doc = NO_MORE_DOCS
		return s.doc, nil
	}
	if s.exclDisi == nil {
		s.doc, err = s.reqScorer.Advance(target)
		if err != nil {
			return
		}
	}

	d, err = s.reqScorer.Advance(target)
	if err != nil {
		return
	}
	if d == NO_MORE_DOCS {
		s.reqScorer = nil
		s.doc = NO_MORE_DOCS
		return s.doc, nil
	}
	s.doc, err = s.toNonExcluded()
	return s.doc, err
}

func (s *ReqExclScorer) DocId() int {
	return s.doc
}

/** Returns the score of the current document matching the query.
 * Initially invalid, until {@link #nextDoc()} is called the first time.
 * @return The score of the required scorer, eventually increased by the score
 * of the optional scorer when it also matches the current document.
 */
func (s *ReqExclScorer) Score() (reqScore float32, err error) {
	return s.reqScorer.Score()
}

func (s *ReqExclScorer) Freq() (n int, err error) {
	return s.reqScorer.Freq()
}

func (s *ReqExclScorer) Cost() int64 {
	return s.reqScorer.Cost()
}
