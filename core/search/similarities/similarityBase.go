package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
	"math"
)

type SimilarityBase interface {
	Similarity
	newStats(string, float32) Stats
	fillStats(Stats, search.CollectionStatistics, search.TermStatistics)
	score(Stats, float32, float32) float32
	explain(search.ExplanationSPI, Stats, int, float32, float32)
	explainScore(Stats, int, search.Explanation, float32) search.Explanation
	encodeNormValue(float32, int) byte
	decodeNormValue(byte) float32
	String() string
}

// internal
type similarityBaseSPI interface {
	newStats(string, float32) Stats
	fillStats(stats Stats, collectionStats search.CollectionStatistics, termStats search.TermStatistics)
	score(Stats, float32, float32) float32
	explain(search.ExplanationSPI, Stats, int, float32, float32)
	explainScore(Stats, int, search.Explanation, float32) search.Explanation
}

type similarityBaseImpl struct {
	similarityImpl
	owner            similarityBaseSPI
	discountOverlaps bool
}

func NewSimilarityBase(owner similarityBaseSPI) *similarityBaseImpl {
	ans := &similarityBaseImpl{owner: owner, discountOverlaps: true}
	once.Do(func() {
		norm_table = ans.buildNormTable()
	})
	return ans
}

func (sb *similarityBaseImpl) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) SimWeight {

	stats := make([]SimWeight, len(termStats))

	for i := 0; i < len(termStats); i++ {
		stats[i] = sb.owner.newStats(collectionStats.Field(), queryBoost)
		sb.owner.fillStats(stats[i].(Stats), collectionStats, termStats[i])
	}

	if len(stats) == 1 {
		return stats[0]
	}

	return newMultiStats(stats)
}

func (sb *similarityBaseImpl) newStats(field string, queryBoost float32) Stats {
	return NewBasicStats(field, queryBoost)
}

func (sb *similarityBaseImpl) fillStats(stats Stats, collectionStats search.CollectionStatistics, termStats search.TermStatistics) {
	assert(collectionStats.SumTotalTermFreq() == -1 || collectionStats.SumTotalTermFreq() >= termStats.TotalTermFreq)

	numberOfDocuments := collectionStats.MaxDoc()

	docFreq := termStats.DocFreq
	totalTermFreq := termStats.TotalTermFreq

	// codec does not supply totalTermFreq: substitute docFreq
	if totalTermFreq == -1 {
		totalTermFreq = docFreq
	}

	var numberOfFieldTokens int64
	var avgFieldLength float32

	sumTotalTermFreq := collectionStats.SumTotalTermFreq()

	if sumTotalTermFreq <= 0 {
		// field does not exist;
		// We have to provide something if codec doesnt supply these measures,
		// or if someone omitted frequencies for the field... negative values cause
		// NaN/Inf for some scorers.
		numberOfFieldTokens = docFreq
		avgFieldLength = 1
	} else {
		numberOfFieldTokens = sumTotalTermFreq
		avgFieldLength = float32(numberOfFieldTokens) / float32(numberOfDocuments)
	}

	// TODO: add sumDocFreq for field (numberOfFieldPostings)
	stats.SetNumberOfDocuments(numberOfDocuments)
	stats.SetNumberOfFieldTokens(numberOfFieldTokens)
	stats.SetAvgFieldLength(avgFieldLength)
	stats.SetDocFreq(docFreq)
	stats.SetTotalTermFreq(totalTermFreq)
}

func (dfr *similarityBaseImpl) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
}

func (sb *similarityBaseImpl) explainScore(stats Stats, doc int, freq search.Explanation, docLen float32) search.Explanation {
	result := search.NewExplanation(sb.owner.score(stats, freq.Value(), docLen), fmt.Sprintf("score(doc=%v, freq=%v), computed from:", doc, freq.Value()))
	result.AddDetail(freq)

	sb.owner.explain(result, stats, doc, freq.Value(), docLen)

	return result
}

func (sb *similarityBaseImpl) SimScorer(stats SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	if v, ok := stats.(*multiStats); ok {
		// a multi term query (e.g. phrase). return the summation,
		// scoring almost as if it were boolean query
		subStats := v.subStats
		subScorers := make([]SimScorer, len(subStats))
		for i := 0; i < len(subScorers); i++ {
			basicstats := subStats[i].(Stats)
			ndv, err := ctx.Reader().(index.AtomicReader).NormValues(basicstats.Field())
			if err != nil {
				return nil, err
			}
			subScorers[i] = newBasicSimScorer(sb, basicstats, ndv)
		}
		return newMultiSimScorer(subScorers), nil
	}

	basicstats := stats.(Stats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(basicstats.Field())
	if err != nil {
		return nil, err
	}

	return newBasicSimScorer(sb, basicstats, ndv), nil
}

func (sb *similarityBaseImpl) buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		f := util.Byte315ToFloat(byte(i))
		table[i] = 1.0 / (f * f)
	}
	return table
}

func (sb *similarityBaseImpl) ComputeNorm(state *index.FieldInvertState) int64 {
	var numTerms int
	if sb.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}

	return int64(sb.encodeNormValue(state.Boost(), numTerms))
}

func (sb *similarityBaseImpl) encodeNormValue(boost float32, fieldLength int) byte {
	return byte(util.FloatToByte315(boost / float32(math.Sqrt(float64(fieldLength)))))
}

func (sb *similarityBaseImpl) decodeNormValue(b byte) float32 {
	return norm_table[b&0xFF]
}

type basicSimScorer struct {
	owner *similarityBaseImpl
	stats Stats
	norms spi.NumericDocValues
}

func newBasicSimScorer(owner *similarityBaseImpl, stats Stats, norms spi.NumericDocValues) *basicSimScorer {
	return &basicSimScorer{owner, stats, norms}
}

func (ss *basicSimScorer) Score(doc int, freq float32) float32 {
	var norm float32

	if ss.norms == nil {
		norm = 1
	} else {
		norm = ss.owner.decodeNormValue(byte(ss.norms(doc)))
	}

	return ss.owner.owner.score(ss.stats, freq, norm)
}

func (ss *basicSimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	var norm float32

	if ss.norms == nil {
		norm = 1
	} else {
		norm = ss.owner.decodeNormValue(byte(ss.norms(doc)))
	}

	return ss.owner.owner.explainScore(ss.stats, doc, freq, norm)
}

func (ss *basicSimScorer) ComputeSlopFactor(distance int) float32 {
	return float32(1) / (float32(distance) + 1)
}

func (ss *basicSimScorer) ComputePayloadFactor(doc, start, end int, payload *util.BytesRef) float32 {
	return 1
}
