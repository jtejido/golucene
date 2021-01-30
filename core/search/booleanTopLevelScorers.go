package search

import (
  . "github.com/jtejido/golucene/core/search/model"
)

type BoostedScorer struct {
  *FilterScorer
  boost float32
}

/**
 * Used when there is more than one scorer in a query, but a segment
 * only had one non-null scorer. This just wraps that scorer directly
 * to factor in coord().
 */
func newBoostedScorer(in Scorer, boost float32) (*BoostedScorer, error) {
  ans := &BoostedScorer{
    boost: boost,
  }
  ans.FilterScorer = newFilterScorer(in)
  return ans, nil
}

func (s *BoostedScorer) Score() (sc float32, err error) {
  sc, err = s.in.Score()
  return sc * s.boost, err
}

/**
 * Used when there are both mandatory and optional clauses, but minShouldMatch
 * dictates that some of the optional clauses must match. The query is a conjunction,
 * but must compute coord based on how many optional subscorers matched (freq).
 */
type CoordinatingConjunctionScorer struct {
  *ConjunctionScorer
  coords   []float32
  reqCount int
  req, opt Scorer
}

func newCoordinatingConjunctionScorer(weight Weight, coords []float32, req Scorer, reqCount int, opt Scorer) (ccs *CoordinatingConjunctionScorer, err error) {
  ans := &CoordinatingConjunctionScorer{
    coords:   coords,
    req:      req,
    reqCount: reqCount,
    opt:      opt,
  }
  ans.ConjunctionScorer, err = newConjunctionScorer(weight, []Scorer{req, opt})
  return ans, err
}

func (s *CoordinatingConjunctionScorer) Score() (sc float32, err error) {
  var rs, os float32
  rs, err = s.req.Score()
  if err != nil {
    return
  }

  os, err = s.opt.Score()
  if err != nil {
    return
  }

  var of int
  of, err = s.opt.Freq()
  if err != nil {
    return
  }
  return (rs + os) * s.coords[s.reqCount+of], nil
}

/**
 * Used when there are mandatory clauses with one optional clause: we compute
 * coord based on whether the optional clause matched or not.
 */
type ReqSingleOptScorer struct {
  *ReqOptSumScorer
  coordReq, coordBoth float32
}

func newReqSingleOptScorer(reqScorer, optScorer Scorer, coordReq, coordBoth float32) (rsos *ReqSingleOptScorer, err error) {
  ans := &ReqSingleOptScorer{
    coordReq:  coordReq,
    coordBoth: coordBoth,
  }
  ans.ReqOptSumScorer, err = newReqOptSumScorer(reqScorer, optScorer)
  return ans, err
}

func (s *ReqSingleOptScorer) Score() (reqScore float32, err error) {
  curDoc := s.reqScorer.DocId()
  reqScore, err = s.reqScorer.Score()
  if err != nil {
    return
  }
  if s.optScorer == nil {
    reqScore *= s.coordReq
    return
  }

  optScorerDoc := s.optScorer.DocId()
  optScorerDoc, err = s.optScorer.Advance(curDoc)
  if err != nil {
    return
  }
  if optScorerDoc < curDoc && optScorerDoc == NO_MORE_DOCS {
    s.optScorer = nil
    reqScore *= s.coordReq
    return
  }

  if optScorerDoc == curDoc {
    var sc float32
    sc, err = s.optScorer.Score()
    if err != nil {
      return
    }
    reqScore += sc
    reqScore *= s.coordBoth
  } else {
    reqScore *= s.coordReq
  }

  return
}

/**
 * Used when there are mandatory clauses with optional clauses: we compute
 * coord based on how many optional subscorers matched (freq).
 */
type ReqMultiOptScorer struct {
  *ReqOptSumScorer
  requiredCount int
  coords        []float32
}

func newReqMultiOptScorer(reqScorer, optScorer Scorer, requiredCount int, coords []float32) (rmos *ReqMultiOptScorer, err error) {
  ans := &ReqMultiOptScorer{
    requiredCount: requiredCount,
    coords:        coords,
  }
  ans.ReqOptSumScorer, err = newReqOptSumScorer(reqScorer, optScorer)
  return ans, err
}

func (s *ReqMultiOptScorer) Score() (reqScore float32, err error) {
  curDoc := s.reqScorer.DocId()
  reqScore, err = s.reqScorer.Score()
  if err != nil {
    return
  }
  if s.optScorer == nil {
    reqScore *= s.coords[s.requiredCount]
    return
  }

  optScorerDoc := s.optScorer.DocId()
  optScorerDoc, err = s.optScorer.Advance(curDoc)
  if err != nil {
    return
  }
  if optScorerDoc < curDoc && optScorerDoc == NO_MORE_DOCS {
    s.optScorer = nil
    reqScore *= s.coords[s.requiredCount]
    return
  }

  if optScorerDoc == curDoc {
    var sc float32
    sc, err = s.optScorer.Score()
    if err != nil {
      return
    }
    reqScore += sc
    var of int
    of, err = s.optScorer.Freq()
    if err != nil {
      return
    }
    reqScore *= s.coords[s.requiredCount+of]
  } else {
    reqScore *= s.coords[s.requiredCount]
  }

  return
}
