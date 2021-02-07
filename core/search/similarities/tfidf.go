package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
)

type TFIDFSimilarity interface {
	Similarity
	tf(float32) float32
	idf(docFreq, numDocs int64) float32
	lengthNorm(*index.FieldInvertState) float32
	encodeNormValue(float32) int64
	decodeNormValue(int64) float32
	sloppyFreq(int) float32
	scorePayload(doc, start, end int, payload *util.BytesRef) float32
}

type tfIdfSimilaritySPI interface {
	tf(float32) float32
	idf(docFreq, numDocs int64) float32
	lengthNorm(*index.FieldInvertState) float32
	encodeNormValue(float32) int64
	decodeNormValue(int64) float32
	sloppyFreq(int) float32
	scorePayload(doc, start, end int, payload *util.BytesRef) float32
}

type TFIDFSimilarityImpl struct {
	similarityImpl
	owner tfIdfSimilaritySPI
}

func NewTFIDFSimilarity(owner TFIDFSimilarity) *TFIDFSimilarityImpl {
	return &TFIDFSimilarityImpl{owner: owner}
}

func (ts *TFIDFSimilarityImpl) idfExplainTerm(collectionStats search.CollectionStatistics, termStats search.TermStatistics) search.Explanation {
	df, max := termStats.DocFreq, collectionStats.MaxDoc()
	idf := ts.owner.idf(df, max)
	return search.NewExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (ts *TFIDFSimilarityImpl) idfExplainPhrase(collectionStats search.CollectionStatistics, termStats []search.TermStatistics) search.Explanation {
	details := make([]search.Explanation, len(termStats))
	var idf float32 = 0

	for i, stat := range termStats {
		details[i] = ts.idfExplainTerm(collectionStats, stat)
		idf += details[i].(*search.ExplanationImpl).Value()
	}
	ans := search.NewExplanation(idf, fmt.Sprintf("idf(), sum of:"))
	ans.SetDetails(details)
	return ans
}

func (ts *TFIDFSimilarityImpl) ComputeNorm(state *index.FieldInvertState) int64 {
	return ts.owner.encodeNormValue(ts.owner.lengthNorm(state))
}

func (ts *TFIDFSimilarityImpl) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) SimWeight {
	var idf search.Explanation
	if len(termStats) == 1 {
		idf = ts.idfExplainTerm(collectionStats, termStats[0])
	} else {
		idf = ts.idfExplainPhrase(collectionStats, termStats)
	}
	return newIDFStats(collectionStats.Field(), idf, queryBoost)
}

func (ts *TFIDFSimilarityImpl) SimScorer(stats SimWeight, ctx *index.AtomicReaderContext) (ss SimScorer, err error) {
	idfstats := stats.(*idfStats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(idfstats.field)
	if err != nil {
		return nil, err
	}

	return newTFIDFSimScorer(ts, idfstats, ndv), nil
}

type tfIDFSimScorer struct {
	owner       *TFIDFSimilarityImpl
	stats       *idfStats
	weightValue float32
	norms       spi.NumericDocValues
}

func newTFIDFSimScorer(owner *TFIDFSimilarityImpl, stats *idfStats, norms spi.NumericDocValues) *tfIDFSimScorer {
	return &tfIDFSimScorer{owner: owner, stats: stats, weightValue: stats.value, norms: norms}
}

func (ss *tfIDFSimScorer) Score(doc int, freq float32) float32 {
	raw := ss.owner.owner.tf(freq) * ss.weightValue // compute tf(f)*weight
	if ss.norms == nil {
		return raw
	}
	return raw * ss.owner.owner.decodeNormValue(ss.norms(doc)) // normalize for field
}

func (ss *tfIDFSimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	return ss.owner.explainScore(doc, freq, ss.stats, ss.norms)
}

func (ss *tfIDFSimScorer) ComputeSlopFactor(distance int) float32 {
	return ss.owner.owner.sloppyFreq(distance)
}

func (ss *tfIDFSimScorer) ComputePayloadFactor(doc, start, end int, payload *util.BytesRef) float32 {
	return ss.owner.owner.scorePayload(doc, start, end, payload)
}

/** Collection statistics for the TF-IDF model. The only statistic of interest
 * to this model is idf. */
type idfStats struct {
	field string
	/** The idf and its explanation */
	idf         search.Explanation
	queryNorm   float32
	queryWeight float32
	queryBoost  float32
	value       float32
}

func newIDFStats(field string, idf search.Explanation, queryBoost float32) *idfStats {
	// TODO: validate?
	return &idfStats{
		field:       field,
		idf:         idf,
		queryBoost:  queryBoost,
		queryWeight: idf.Value() * queryBoost, // compute query weight
	}
}

func (stats *idfStats) ValueForNormalization() float32 {
	// TODO: (sorta LUCENE-1907) make non-static class and expose this squaring via a nice method to subclasses?
	return stats.queryWeight * stats.queryWeight // sum of squared weights
}

func (stats *idfStats) Normalize(queryNorm float32, topLevelBoost float32) {
	stats.queryNorm = queryNorm * topLevelBoost
	stats.queryWeight *= stats.queryNorm                                          // normalize query weight
	stats.value = stats.queryWeight * stats.idf.(*search.ExplanationImpl).Value() // idf for document
}

func (stats *idfStats) Field() string {
	return stats.field
}

func (ss *TFIDFSimilarityImpl) explainScore(doc int, freq search.Explanation,
	stats *idfStats, norms spi.NumericDocValues) search.Explanation {

	// explain query weight
	boostExpl := search.NewExplanation(stats.queryBoost, "boost")
	queryNormExpl := search.NewExplanation(stats.queryNorm, "queryNorm")
	queryExpl := search.NewExplanation(
		boostExpl.Value()*stats.idf.Value()*queryNormExpl.Value(),
		"queryWeight, product of:")
	if stats.queryBoost != 1 {
		queryExpl.AddDetail(boostExpl)
	}
	queryExpl.AddDetail(stats.idf)
	queryExpl.AddDetail(queryNormExpl)

	// explain field weight
	tfExplanation := search.NewExplanation(ss.owner.tf(freq.Value()),
		fmt.Sprintf("tf(freq=%v), with freq of:", freq.Value()))
	tfExplanation.AddDetail(freq)
	fieldNorm := float32(1)
	if norms != nil {
		fieldNorm = ss.owner.decodeNormValue(norms(doc))
	}
	fieldNormExpl := search.NewExplanation(fieldNorm, fmt.Sprintf("fieldNorm(doc=%v)", doc))
	fieldExpl := search.NewExplanation(
		tfExplanation.Value()*stats.idf.Value()*fieldNormExpl.Value(),
		fmt.Sprintf("fieldWeight in %v, product of:", doc))
	fieldExpl.AddDetail(tfExplanation)
	fieldExpl.AddDetail(stats.idf)
	fieldExpl.AddDetail(fieldNormExpl)

	if queryExpl.Value() == 1 {
		return fieldExpl
	}

	// combine them
	ans := search.NewExplanation(queryExpl.Value()*fieldExpl.Value(),
		fmt.Sprintf("score(doc=%v,freq=%v), product of:", doc, freq))
	ans.AddDetail(queryExpl)
	ans.AddDetail(fieldExpl)
	return ans
}
