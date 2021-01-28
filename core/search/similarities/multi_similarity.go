package similarities

import (
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
)

type MultiSimilarity struct {
	spi  search.Similarity
	sims []search.Similarity
}

func newMultiSimilarity(spi search.Similarity, sims []search.Similarity) *MultiSimilarity {
	ans := &MultiSimilarity{spi: spi, sims: sims}
	return ans
}

func (ms *MultiSimilarity) Coord(overlap, maxOverlap int) float32 {
	return 1.
}

func (ms *MultiSimilarity) QueryNorm(sumOfSquaredWeights float32) float32 {
	return 1.
}

func (ms *MultiSimilarity) ComputeNorm(state *index.FieldInvertState) int64 {
	return ms.sims[0].ComputeNorm(state)
}

func (ms *MultiSimilarity) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) search.SimWeight {
	subStats := make([]search.SimWeight, len(ms.sims))
	for i := 0; i < len(subStats); i++ {
		subStats[i] = ms.sims[i].ComputeWeight(queryBoost, collectionStats, termStats...)
	}
	return newMultiStats(subStats)
}

func (ms *MultiSimilarity) SimScorer(stats search.SimWeight, ctx *index.AtomicReaderContext) (ss search.SimScorer, err error) {
	subScorers := make([]search.SimScorer, len(ms.sims))
	for i := 0; i < len(subScorers); i++ {
		subScorers[i], err = ms.sims[i].SimScorer(stats.(*multiStats).subStats[i], ctx)

		if err != nil {
			return
		}
	}
	return newMultiSimScorer(ms, subScorers), nil
}

type multiSimScorer struct {
	owner      search.Similarity
	subScorers []search.SimScorer
}

func newMultiSimScorer(owner search.Similarity, subScorers []search.SimScorer) *multiSimScorer {
	return &multiSimScorer{owner, subScorers}
}

func (ss *multiSimScorer) Score(doc int, freq float32) float32 {
	var sum float32
	for _, subScorer := range ss.subScorers {
		sum += subScorer.Score(doc, freq)
	}
	return sum
}

func (ss *multiSimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	expl := search.NewExplanation(ss.Score(doc, freq.Value()), "sum of:")
	for _, subScorer := range ss.subScorers {
		expl.AddDetail(subScorer.Explain(doc, freq))
	}
	return expl
}

type multiStats struct {
	subStats []search.SimWeight
}

func newMultiStats(subStats []search.SimWeight) *multiStats {
	return &multiStats{subStats}
}

func (stats *multiStats) ValueForNormalization() float32 {
	var sum float32
	for _, stat := range stats.subStats {
		sum += stat.ValueForNormalization()
	}
	return sum / float32(len(stats.subStats))
}

func (stats *multiStats) Normalize(queryNorm float32, topLevelBoost float32) {
	for _, stat := range stats.subStats {
		stat.Normalize(queryNorm, topLevelBoost)
	}
}

func (stats *multiStats) Field() string {
	return ""
}
