package search

import (
	"fmt"
	. "github.com/jtejido/golucene/core/index/model"
	// . "github.com/jtejido/golucene/core/search/model"
	"math"
)

// search/Scorer.java
type Scorer interface {
	DocsEnum
	Score() (float32, error)
	Weight() Weight
}

type IScorer interface {
	Score() (float32, error)
}

type ScorerSPI interface {
	DocId() int
	NextDoc() (int, error)
}

type abstractScorer struct {
	weight Weight
}

func (s *abstractScorer) Weight() Weight {
	return s.weight
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

// search/BulkScorer.java

/*
This class is used to score a range of documents at once, and is
returned by Weight.BulkScorer(). Only queries that have a more
optimized means of scoring across a range of documents need to
override this. Otherwise, a default implementation is wrapped around
the Scorer returned by Weight.Scorer().
*/
type BulkScorer interface {
	ScoreAndCollect(Collector) error
	ScoreAndCollectUpto(Collector, int) (bool, error)
}

type bulkScorerImplSPI interface {
	ScoreAndCollectUpto(Collector, int) (bool, error)
}

type BulkScorerImpl struct {
	spi bulkScorerImplSPI
}

func newBulkScorer(spi bulkScorerImplSPI) *BulkScorerImpl {
	return &BulkScorerImpl{spi}
}

func (bs *BulkScorerImpl) ScoreAndCollect(collector Collector) (err error) {
	assert(bs != nil)
	assert(bs.spi != nil)
	_, err = bs.spi.ScoreAndCollectUpto(collector, math.MaxInt32)
	return
}
