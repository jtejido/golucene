package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
	"math"
)

const (
	DEFAULT_K1 float32 = 1.2
	DEFAULT_B  float32 = .75
)

type IProbabSimilarity interface {
	IIDF
	score(freq, norm float32) float32
	String() string
}

/**
* The probabilistic relevance model was devised by Robertson and Jones as a framework for probabilistic models to come.
* It is a formalism of information retrieval useful to derive ranking functions used by search engines and web search engines
* in order to rank matching documents according to their relevance to a given search query.

* It makes an estimation of the probability of finding if a document dj is relevant to a query q.
* This model assumes that this probability of relevance depends on the query and document representations.
* Furthermore, it assumes that there is a portion of all documents that is preferred by the user as the answer set for query q.
* Such an ideal answer set is called R and should maximize the overall probability of relevance to that user.
* The prediction is that documents in this set R are relevant to the query, while documents not present in the set are non-relevant
* and also it is a theoritical model.
*
* S. E. Robertson and K. S. Jones (May–June 1976), Relevance weighting of search terms,
* Journal of the American Society for Information Science, pp. 129–146
*
* Stephen Robertson and Hugo Zaragoza (2009). "The Probabilistic Relevance Framework: BM25 and Beyond".
* Trends Inf. Retr.: 333–389.
*
* NOTE: This is the base in which BM25 and its family extends. It should have room for others like it. (BIM, etc.)
*
**/
type ProbabilitySimilarity struct {
	SimilarityBase
	spi              IProbabSimilarity
	k1, b            float32
	discountOverlaps bool
}

func newProbabilitySimilarity(spi IProbabSimilarity, k1, b float32, discountOverlaps bool) *ProbabilitySimilarity {
	ans := &ProbabilitySimilarity{spi: spi, k1: k1, b: b, discountOverlaps: discountOverlaps}
	NORM_TABLE = ans.buildNormTable()
	return ans
}

func (bbm25 *ProbabilitySimilarity) buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		f := util.Byte315ToFloat(byte(i))
		table[i] = 1.0 / (f * f)
	}
	return table
}

func (bbm25 *ProbabilitySimilarity) idf(docFreq, numDocs int64) float32 {
	return bbm25.spi.idf(docFreq, numDocs)
}

func (bbm25 *ProbabilitySimilarity) avgFieldLength(collectionStats search.CollectionStatistics) float32 {
	sumTotalTermFreq := collectionStats.SumTotalTermFreq()

	if sumTotalTermFreq <= 0 {
		return 1.
	}

	return (float32(sumTotalTermFreq) / float32(collectionStats.MaxDoc()))
}

func (bbm25 *ProbabilitySimilarity) encodeNormValue(boost float32, fieldLength int) byte {
	return byte(util.FloatToByte315(boost / float32(math.Sqrt(float64(fieldLength)))))
}

func (bbm25 *ProbabilitySimilarity) decodeNormValue(b byte) float32 {
	return NORM_TABLE[b&0xFF]
}

func (bbm25 *ProbabilitySimilarity) idfExplainTerm(collectionStats search.CollectionStatistics, termStats search.TermStatistics) search.Explanation {
	df, max := termStats.DocFreq, collectionStats.MaxDoc()
	idf := bbm25.idf(df, max)
	return search.NewExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (bbm25 *ProbabilitySimilarity) idfExplainPhrase(collectionStats search.CollectionStatistics, termStats []search.TermStatistics) search.Explanation {
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

func (bbm25 *ProbabilitySimilarity) ComputeNorm(state *index.FieldInvertState) int64 {
	var numTerms int
	if bbm25.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}

	return int64(bbm25.encodeNormValue(state.Boost(), numTerms))
}

func (bbm25 *ProbabilitySimilarity) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) search.SimWeight {
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

func (bbm25 *ProbabilitySimilarity) score(freq, norm float32) float32 {
	return bbm25.spi.score(freq, norm)
}

func (bbm25 *ProbabilitySimilarity) explainScore(doc int, freq search.Explanation,
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

func (bbm25 *ProbabilitySimilarity) SimScorer(stats search.SimWeight, ctx *index.AtomicReaderContext) (ss search.SimScorer, err error) {
	bm25Stats := stats.(*bm25Stats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(bm25Stats.field)
	if err != nil {
		return nil, err
	}
	return newBM25SimScorer(bbm25, bm25Stats, ndv), nil
}

func (bbm25 *ProbabilitySimilarity) String() string {
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
	owner       *ProbabilitySimilarity
	stats       *bm25Stats
	weightValue float32
	norms       spi.NumericDocValues
	cache       []float32
}

func newBM25SimScorer(owner *ProbabilitySimilarity, stats *bm25Stats, norms spi.NumericDocValues) *bm25SimScorer {
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

func (stats *bm25Stats) Field() string {
	return stats.field
}
