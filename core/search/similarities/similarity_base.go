package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
	"math"
	"reflect"
)

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

type IIDF interface {
	/** Computes a score factor based on a term's document frequency (the number
	 * of documents which contain the term).  This value is multiplied by the
	 * {@link #tf(float)} factor for each term in the query and these products are
	 * then summed to form the initial score for a document.
	 *
	 * <p>Terms that occur in fewer documents are better indicators of topic, so
	 * implementations of this method usually return larger values for rare terms,
	 * and smaller values for common terms.
	 *
	 * @param docFreq the number of documents which contain the term
	 * @param numDocs the total number of documents in the collection
	 * @return a score factor based on the term's document frequency
	 */
	idf(docFreq int64, numDocs int64) float32
}

type BaseSimilarity interface {
	IIDF
	// Compute an index-time normalization value for this field instance.
	//
	// This value will be stored in a single byte lossy representation
	// by encodeNormValue().
	lengthNorm(*index.FieldInvertState) float32
	// Decodes a normalization factor stored in an index.
	decodeNormValue(norm int64) float32
	// Encodes a normalization factor for storage in an index.
	encodeNormValue(float32) int64
}

type SimilarityBase struct {
	spi              IBaseLMSimilarity
	discountOverlaps bool
}

func newSimilarityBase(spi IBaseLMSimilarity) *SimilarityBase {
	ans := &SimilarityBase{spi: spi, discountOverlaps: true}
	NORM_TABLE = ans.buildNormTable()
	return ans
}

func (b *SimilarityBase) buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		f := util.Byte315ToFloat(byte(i))
		table[i] = 1.0 / (f * f)
	}
	return table
}

func (b *SimilarityBase) Coord(overlap, maxOverlap int) float32 {
	return 1.
}

func (b *SimilarityBase) QueryNorm(sumOfSquaredWeights float32) float32 {
	return 1.
}

func (b *SimilarityBase) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) search.SimWeight {

	stats := make([]search.SimWeight, len(termStats))

	for i := 0; i < len(termStats); i++ {
		stats[i] = b.newStats(collectionStats.Field(), queryBoost)
		b.fillBasicStats(stats[i], collectionStats, termStats[i])
	}

	if len(stats) == 1 {
		return stats[0]
	}

	return newMultiStats(stats)
}

func (b *SimilarityBase) ComputeNorm(state *index.FieldInvertState) int64 {
	var numTerms int
	if b.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}

	return int64(b.encodeNormValue(state.Boost(), numTerms))
}

func (b *SimilarityBase) newStats(field string, queryBoost float32) IBasicStats {
	return b.spi.newStats(field, queryBoost)
}

func (b *SimilarityBase) fillBasicStats(sStats search.SimWeight, collectionStats search.CollectionStatistics, termStats search.TermStatistics) {
	stats := sStats.(IBasicStats)
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

	b.spi.fillBasicStats(stats, collectionStats, termStats)
}

func (b *SimilarityBase) score(stats IBasicStats, freq, docLen float32) float32 {
	return b.spi.score(stats, freq, docLen)
}

func (b *SimilarityBase) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	b.spi.explain(expl, stats, doc, freq, docLen)
}

func (b *SimilarityBase) explainScore(stats IBasicStats, doc int, freq search.Explanation, docLen float32) search.Explanation {
	result := search.NewExplanation(b.score(stats, freq.Value(), docLen), fmt.Sprintf("score(doc=%v, freq=%v), computed from:", doc, freq.Value()))
	result.AddDetail(freq)

	b.explain(result, stats, doc, freq.Value(), docLen)

	return result
}

func (b *SimilarityBase) SimScorer(stats search.SimWeight, ctx *index.AtomicReaderContext) (ss search.SimScorer, err error) {

	if reflect.TypeOf(stats) == reflect.TypeOf((*multiStats)(nil)) {
		// a multi term query (e.g. phrase). return the summation,
		// scoring almost as if it were boolean query
		subStats := stats.(*multiStats).subStats
		subScorers := make([]search.SimScorer, len(subStats))
		for i := 0; i < len(subScorers); i++ {
			basicstats := subStats[i].(IBasicStats)
			ndv, err := ctx.Reader().(index.AtomicReader).NormValues(basicstats.Field())
			if err != nil {
				return nil, err
			}
			subScorers[i] = newBasicSimScorer(b, basicstats, ndv)
		}
		return newMultiSimScorer(b, subScorers), nil
	} else {
		bstats := stats.(IBasicStats)
		ndv, err := ctx.Reader().(index.AtomicReader).NormValues(bstats.Field())
		if err != nil {
			return nil, err
		}
		return newBasicSimScorer(b, bstats, ndv), nil
	}
}

func (b *SimilarityBase) encodeNormValue(boost float32, fieldLength int) byte {
	return byte(util.FloatToByte315(boost / float32(math.Sqrt(float64(fieldLength)))))
}

func (b *SimilarityBase) decodeNormValue(bb byte) float32 {
	return NORM_TABLE[bb&0xFF]
}

func (b *SimilarityBase) String() string {
	return b.spi.String()
}

type basicSimScorer struct {
	owner *SimilarityBase
	stats IBasicStats
	norms spi.NumericDocValues
}

func newBasicSimScorer(owner *SimilarityBase, stats IBasicStats, norms spi.NumericDocValues) *basicSimScorer {
	return &basicSimScorer{owner, stats, norms}
}

func (ss *basicSimScorer) Score(doc int, freq float32) float32 {
	var norm float32

	if ss.norms == nil {
		norm = 1
	} else {
		norm = ss.owner.decodeNormValue(byte(ss.norms(doc)))
	}

	return ss.owner.score(ss.stats, freq, norm)
}

func (ss *basicSimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	var norm float32

	if ss.norms == nil {
		norm = 1
	} else {
		norm = ss.owner.decodeNormValue(byte(ss.norms(doc)))
	}

	return ss.owner.explainScore(ss.stats, doc, freq, norm)
}
