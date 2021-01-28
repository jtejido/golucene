package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/codec/spi"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
)

// search/similarities/TFIDFSimilarity.java

type ITFIDFSimilarity interface {
	BaseSimilarity
	/** Computes a score factor based on a term or phrase's frequency in a
	 * document.  This value is multiplied by the {@link #idf(long, long)}
	 * factor for each term in the query and these products are then summed to
	 * form the initial score for a document.
	 *
	 * <p>Terms and phrases repeated in a document indicate the topic of the
	 * document, so implementations of this method usually return larger values
	 * when <code>freq</code> is large, and smaller values when <code>freq</code>
	 * is small.
	 *
	 * @param freq the frequency of a term within a document
	 * @return a score factor based on a term's within-document frequency
	 */
	tf(freq float32) float32
}

type TFIDFSimilarity struct {
	SimilarityBase
	spi ITFIDFSimilarity
}

func newTFIDFSimilarity(spi ITFIDFSimilarity) *TFIDFSimilarity {
	return &TFIDFSimilarity{spi: spi}
}

func (ts *TFIDFSimilarity) idfExplainTerm(collectionStats search.CollectionStatistics, termStats search.TermStatistics) search.Explanation {
	df, max := termStats.DocFreq, collectionStats.MaxDoc()
	idf := ts.spi.idf(df, max)
	return search.NewExplanation(idf, fmt.Sprintf("idf(docFreq=%v, maxDocs=%v)", df, max))
}

func (ts *TFIDFSimilarity) idfExplainPhrase(collectionStats search.CollectionStatistics, termStats []search.TermStatistics) search.Explanation {
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

func (ts *TFIDFSimilarity) ComputeNorm(state *index.FieldInvertState) int64 {
	return ts.spi.encodeNormValue(ts.spi.lengthNorm(state))
}

func (ts *TFIDFSimilarity) ComputeWeight(queryBoost float32, collectionStats search.CollectionStatistics, termStats ...search.TermStatistics) search.SimWeight {
	var idf search.Explanation
	if len(termStats) == 1 {
		idf = ts.idfExplainTerm(collectionStats, termStats[0])
	} else {
		idf = ts.idfExplainPhrase(collectionStats, termStats)
	}
	return newIDFStats(collectionStats.Field(), idf, queryBoost)
}

func (ts *TFIDFSimilarity) SimScorer(stats search.SimWeight, ctx *index.AtomicReaderContext) (ss search.SimScorer, err error) {
	idfstats := stats.(*idfStats)
	ndv, err := ctx.Reader().(index.AtomicReader).NormValues(idfstats.field)
	if err != nil {
		return nil, err
	}
	return newTFIDFSimScorer(ts, idfstats, ndv), nil
}

type tfIDFSimScorer struct {
	owner       *TFIDFSimilarity
	stats       *idfStats
	weightValue float32
	norms       spi.NumericDocValues
}

func newTFIDFSimScorer(owner *TFIDFSimilarity, stats *idfStats, norms spi.NumericDocValues) *tfIDFSimScorer {
	return &tfIDFSimScorer{owner, stats, stats.value, norms}
}

func (ss *tfIDFSimScorer) Score(doc int, freq float32) float32 {
	raw := ss.owner.spi.tf(freq) * ss.weightValue // compute tf(f)*weight
	if ss.norms == nil {
		return raw
	}
	return raw * ss.owner.spi.decodeNormValue(ss.norms(doc)) // normalize for field
}

func (ss *tfIDFSimScorer) Explain(doc int, freq search.Explanation) search.Explanation {
	return ss.owner.explainScore(doc, freq, ss.stats, ss.norms)
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
		queryWeight: idf.(*search.ExplanationImpl).Value() * queryBoost, // compute query weight
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

func (ss *TFIDFSimilarity) explainScore(doc int, freq search.Explanation,
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
	tfExplanation := search.NewExplanation(ss.spi.tf(freq.Value()),
		fmt.Sprintf("tf(freq=%v), with freq of:", freq.Value()))
	tfExplanation.AddDetail(freq)
	fieldNorm := float32(1)
	if norms != nil {
		fieldNorm = ss.spi.decodeNormValue(norms(doc))
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
