package similarities

import (
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
)

type PerFieldSimilarityWrapperSPI interface {
	Get(name string) Similarity
}

/*
Provides the ability to use a different Similarity for different
fields.

Subclasses should implement Get() to return an appropriate Similarity
(for example, using field-specific parameter values) for the field.
*/
type PerFieldSimilarityWrapper struct {
	spi PerFieldSimilarityWrapperSPI
}

func NewPerFieldSimilarityWrapper(spi PerFieldSimilarityWrapperSPI) *PerFieldSimilarityWrapper {
	return &PerFieldSimilarityWrapper{spi: spi}
}

func (wrapper *PerFieldSimilarityWrapper) ComputeNorm(state *index.FieldInvertState) int64 {
	return wrapper.spi.Get(state.Name()).ComputeNorm(state)
}

func (wrapper *PerFieldSimilarityWrapper) computeWeight(queryBoost float32,
	collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) SimWeight {
	sim := wrapper.spi.Get(collectionStats.Field())
	return &PerFieldSimWeight{sim, sim.ComputeWeight(queryBoost, collectionStats, termStats...)}
}

func (wrapper *PerFieldSimilarityWrapper) simScorer(w SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	panic("not implemented yet")
}

type PerFieldSimWeight struct {
	delegate       Similarity
	delegateWeight SimWeight
}

func (w *PerFieldSimWeight) ValueForNormalization() float32 {
	return w.delegateWeight.ValueForNormalization()
}

func (w *PerFieldSimWeight) Normalize(queryNorm, topLevelBoost float32) {
	w.delegateWeight.Normalize(queryNorm, topLevelBoost)
}

func (w *PerFieldSimWeight) Field() string {
	return ""
}
