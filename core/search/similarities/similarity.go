package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
)

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

type Stats interface {
	SimWeight
	NumberOfDocuments() int64
	SetNumberOfDocuments(numberOfDocuments int64)
	NumberOfFieldTokens() int64
	SetNumberOfFieldTokens(numberOfFieldTokens int64)
	AvgFieldLength() float32
	SetAvgFieldLength(avgFieldLength float32)
	DocFreq() int64
	SetDocFreq(docFreq int64)
	TotalTermFreq() int64
	SetTotalTermFreq(totalTermFreq int64)
	TotalBoost() float32
	Field() string
}

type Similarity = search.Similarity

type similarityImpl struct{}

func (b *similarityImpl) Coord(overlap, maxOverlap int) float32 {
	return 1.
}

func (b *similarityImpl) QueryNorm(sumOfSquaredWeights float32) float32 {
	return 1.
}

type SimScorer = search.SimScorer

// internal
type simScorerSPI interface {
	Score(doc int, freq float32) float32
}

type simScorerImpl struct {
	owner simScorerSPI
}

func NewSimScorer(owner simScorerSPI) *simScorerImpl {
	return &simScorerImpl{owner}
}

func (ss *simScorerImpl) Explain(doc int, freq search.Explanation) search.Explanation {
	result := search.NewExplanation(ss.owner.Score(doc, freq.Value()), fmt.Sprintf("score(doc=%d,freq=%.2f), with freq of:", doc, freq.Value()))
	result.AddDetail(freq)
	return result
}

type SimWeight = search.SimWeight
