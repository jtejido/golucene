package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
	"math"
)

var _ Similarity = (*BM25Similarity)(nil)

const (
	DEFAULT_K1 float32 = 1.2
	DEFAULT_B  float32 = .75
)

/**
 * BM25 is a class for ranking documents against a query.
 *
 * The implementation is based on the paper by Stephen E. Robertson, Steve Walker, Susan Jones,
 * Micheline Hancock-Beaulieu & Mike Gatford (November 1994).
 * @see http://trec.nist.gov/pubs/trec3/t3_proceedings.html.
 *
 * Some modifications have been made to allow for non-negative scoring as suggested here.
 * @see https://doc.rero.ch/record/16754/files/Dolamic_Ljiljana_-_When_Stopword_Lists_Make_the_Difference_20091218.pdf
 */
type BM25Similarity struct {
	baseBM25Similarity
}

func NewBM25Similarity(k1, b float32, discountOverlaps bool) *BM25Similarity {
	ans := new(BM25Similarity)
	ans.k1 = k1
	ans.b = b
	ans.discountOverlaps = discountOverlaps
	ans.spi = ans
	once.Do(func() {
		norm_table = ans.buildNormTable()
	})
	return ans
}

func NewDefaultBM25Similarity() *BM25Similarity {
	return NewBM25Similarity(DEFAULT_K1, DEFAULT_B, true)
}

func (bm25 *BM25Similarity) idf(docFreq, numDocs int64) float32 {
	return float32(math.Log(1. + (float64(numDocs)-float64(docFreq)+0.5)/(float64(docFreq)+0.5)))
}

func (bm25 *BM25Similarity) score(freq, norm float32) float32 {
	num := freq * (bm25.k1 + 1)
	denom := freq + bm25.k1*(1-bm25.b+bm25.b*norm)

	return num / denom
}

func (bm25 *BM25Similarity) String() string {
	return fmt.Sprintf("BM25(k1=%v, b=%v)", bm25.k1, bm25.b)
}

type bm25SimilaritySPI interface {
	idf(docFreq, numDocs int64) float32
	score(freq, norm float32) float32
	String() string
}

type baseBM25Similarity struct {
	similarityImpl
	spi              bm25SimilaritySPI
	k1, b            float32
	discountOverlaps bool
}

func (bbm25 *baseBM25Similarity) buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		f := util.Byte315ToFloat(byte(i))
		table[i] = 1.0 / (f * f)
	}
	return table
}

func (bbm25 *baseBM25Similarity) idf(docFreq, numDocs int64) float32 {
	return bbm25.spi.idf(docFreq, numDocs)
}

func (bbm25 *baseBM25Similarity) sloppyFreq(distance int) float32 {
	return 1.0 / (float32(distance) + 1)
}

func (bbm25 *baseBM25Similarity) scorePayload(doc, start, end int, payload *util.BytesRef) float32 {
	return 1
}

func (bbm25 *baseBM25Similarity) avgFieldLength(collectionStats search.CollectionStatistics) float32 {
	sumTotalTermFreq := collectionStats.SumTotalTermFreq()

	if sumTotalTermFreq <= 0 {
		return 1.
	}

	return (float32(sumTotalTermFreq) / float32(collectionStats.MaxDoc()))
}

func (bbm25 *baseBM25Similarity) encodeNormValue(boost float32, fieldLength int) byte {
	return byte(util.FloatToByte315(boost / float32(math.Sqrt(float64(fieldLength)))))
}

func (bbm25 *baseBM25Similarity) decodeNormValue(b byte) float32 {
	return norm_table[b&0xFF]
}

func (bbm25 *baseBM25Similarity) idfExplainTerm(collectionStats search.CollectionStatistics, termStats search.TermStatistics) search.Explanation {
	df, max := termStats.DocFreq, collectionStats.MaxDoc()
	idf := bbm25.idf(df, max)
	return search.NewExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (bbm25 *baseBM25Similarity) idfExplainPhrase(collectionStats search.CollectionStatistics, termStats []search.TermStatistics) search.Explanation {
	details := make([]search.Explanation, len(termStats))
	var idf float32 = 0
	for i, stat := range termStats {
		details[i] = bbm25.idfExplainTerm(collectionStats, stat)
		idf += details[i].(*search.ExplanationImpl).Value()
	}
	ans := search.NewExplanation(idf, fmt.Sprintf("idf(), sum of:"))
	ans.SetDetails(details)
	return ans
}

func (bbm25 *baseBM25Similarity) ComputeNorm(state *index.FieldInvertState) int64 {
	var numTerms int
	if bbm25.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}

	return int64(bbm25.encodeNormValue(state.Boost(), numTerms))
}

func (bbm25 *baseBM25Similarity) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) search.SimWeight {
	var idf search.Explanation
	if len(termStats) == 1 {
		idf = bbm25.idfExplainTerm(collectionStats, termStats[0])
	} else {
		idf = bbm25.idfExplainPhrase(collectionStats, termStats)
	}

	avgdl := bbm25.avgFieldLength(collectionStats)
	cache := make([]float32, 256)
	for i, _ := range cache {
		cache[i] = bbm25.decodeNormValue(byte(i)) / avgdl
	}

	return newBM25Stats(collectionStats.Field(), idf, queryBoost, avgdl, cache)
}

func (bbm25 *baseBM25Similarity) score(freq, norm float32) float32 {
	return bbm25.spi.score(freq, norm)
}

func (bbm25 *baseBM25Similarity) explainScore(doc int, freq search.Explanation,
	stats *bm25Stats, norms spi.NumericDocValues) search.Explanation {

	var tfNormExpl search.ExplanationSPI

	// explain query weight
	boostExpl := search.NewExplanation(stats.queryBoost*stats.topLevelBoost, "boost")

	if norms != nil {
		tfNormExpl = search.NewExplanation((freq.Value()*(bbm25.k1+1))/(freq.Value()+bbm25.k1), "tfNorm, computed from:")
		tfNormExpl.AddDetail(freq)
		tfNormExpl.AddDetail(search.NewExplanation(bbm25.k1, "parameter k1"))
		tfNormExpl.AddDetail(search.NewExplanation(0, "parameter b (norms omitted for field)"))
	} else {
		doclen := bbm25.decodeNormValue(byte(norms(doc)))
		tfNormExpl = search.NewExplanation((freq.Value()*(bbm25.k1+1))/(freq.Value()+bbm25.k1*(1-bbm25.b+bbm25.b*doclen/stats.avgdl)), "tfNorm, computed from:")
		tfNormExpl.AddDetail(search.NewExplanation(bbm25.b, "parameter b"))
		tfNormExpl.AddDetail(search.NewExplanation(stats.avgdl, "avgFieldLength"))
		tfNormExpl.AddDetail(search.NewExplanation(doclen, "fieldLength"))

	}

	// combine them
	ans := search.NewExplanation(boostExpl.Value()*stats.idf.Value()*tfNormExpl.Value(),
		fmt.Sprintf("score(doc=%v,freq=%v), product of:", doc, freq))

	if boostExpl.Value() != 1 {
		ans.AddDetail(boostExpl)
	}

	ans.AddDetail(stats.idf)
	ans.AddDetail(tfNormExpl)

	return ans
}

func (bbm25 *baseBM25Similarity) SimScorer(stats search.SimWeight, ctx *index.AtomicReaderContext) (ss search.SimScorer, err error) {
	bm25Stats := stats.(*bm25Stats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(bm25Stats.field)
	if err != nil {
		return nil, err
	}
	return newBM25SimScorer(bbm25, bm25Stats, ndv), nil
}

func (bbm25 *baseBM25Similarity) String() string {
	return bbm25.spi.String()
}

type bm25Stats struct {
	/** The idf and its explanation */
	idf search.Explanation
	/** The average document length. */
	avgdl float32
	/** query's inner boost */
	queryBoost float32
	/** query's outer boost (only for explain) */
	topLevelBoost float32
	/** weight (idf * boost) */
	weight float32
	/** field name, for pulling norms */
	field string
	/** precomputed norm[256] with k1 * ((1 - b) + b * dl / avgdl) */
	cache []float32
}

type bm25SimScorer struct {
	owner       *baseBM25Similarity
	stats       *bm25Stats
	weightValue float32
	norms       spi.NumericDocValues
	cache       []float32
}

func newBM25SimScorer(owner *baseBM25Similarity, stats *bm25Stats, norms spi.NumericDocValues) *bm25SimScorer {
	return &bm25SimScorer{owner, stats, stats.weight, norms, stats.cache}
}

func (ss *bm25SimScorer) Score(doc int, freq float32) float32 {
	var norm float32

	if ss.norms == nil {
		norm = ss.owner.k1
	} else {
		norm = ss.cache[byte(ss.norms(doc))&0xFF]
	}

	return ss.weightValue * ss.owner.score(freq, norm)
}

func (ss *bm25SimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	return ss.owner.explainScore(doc, freq, ss.stats, ss.norms)
}

func (ss *bm25SimScorer) ComputeSlopFactor(distance int) float32 {
	return ss.owner.sloppyFreq(distance)
}

func (ss *bm25SimScorer) ComputePayloadFactor(doc, start, end int, payload *util.BytesRef) float32 {
	return ss.owner.scorePayload(doc, start, end, payload)
}

func newBM25Stats(field string, idf search.Explanation, queryBoost, avgdl float32, cache []float32) *bm25Stats {
	return &bm25Stats{
		field:      field,
		idf:        idf,
		queryBoost: queryBoost,
		avgdl:      avgdl,
		cache:      cache,
	}
}

func (stats *bm25Stats) ValueForNormalization() float32 {
	queryWeight := stats.idf.(*search.ExplanationImpl).Value() * stats.queryBoost
	return queryWeight * queryWeight
}

func (stats *bm25Stats) Normalize(queryNorm float32, topLevelBoost float32) {
	stats.topLevelBoost = topLevelBoost
	stats.weight = stats.idf.(*search.ExplanationImpl).Value() * stats.queryBoost * topLevelBoost
}
