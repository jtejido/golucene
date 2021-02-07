package similarities

import (
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
)

type MultiSimilarity struct {
	*similarityImpl
	sims []Similarity
}

func NewMultiSimilarity(sims []Similarity) *MultiSimilarity {
	return &MultiSimilarity{sims: sims}
}

func (ms *MultiSimilarity) ComputeNorm(state *index.FieldInvertState) int64 {
	return ms.sims[0].ComputeNorm(state)
}

func (ms *MultiSimilarity) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) SimWeight {
	subStats := make([]SimWeight, len(ms.sims))
	for i := 0; i < len(subStats); i++ {
		subStats[i] = ms.sims[i].ComputeWeight(queryBoost, collectionStats, termStats...)
	}
	return newMultiStats(subStats)
}

func (ms *MultiSimilarity) SimScorer(stats SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	subScorers := make([]SimScorer, len(ms.sims))
	for i := 0; i < len(subScorers); i++ {
		subScorers[i], err = ms.sims[i].SimScorer(stats.(*multiStats).subStats[i], ctx)

		if err != nil {
			return
		}
	}
	return newMultiSimScorer(subScorers), nil
}

type multiSimScorer struct {
	subScorers []SimScorer
}

func newMultiSimScorer(subScorers []SimScorer) *multiSimScorer {
	return &multiSimScorer{subScorers}
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

func (ss *multiSimScorer) ComputeSlopFactor(distance int) float32 {
	return ss.subScorers[0].ComputeSlopFactor(distance)
}

func (ss *multiSimScorer) ComputePayloadFactor(doc, start, end int, payload *util.BytesRef) float32 {
	return ss.subScorers[0].ComputePayloadFactor(doc, start, end, payload)
}

type multiStats struct {
	subStats []SimWeight
}

func newMultiStats(subStats []SimWeight) *multiStats {
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
