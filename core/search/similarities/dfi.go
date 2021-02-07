package similarities

import (
	"fmt"
	"math"
)

var _ Similarity = (*dfiSimilarityImpl)(nil)
var _ DFISimilarity = (*dfiSimilarityImpl)(nil)

type DFISimilarity interface {
	SimilarityBase
}

/**
 * Implements the <em>Divergence from Independence (DFI)</em> model based on Chi-square statistics
 * (i.e., standardized Chi-squared distance from independence in term frequency tf).
 * <p>
 * DFI is both parameter-free and non-parametric:
 * <ul>
 * <li>parameter-free: it does not require any parameter tuning or training.</li>
 * <li>non-parametric: it does not make any assumptions about word frequency distributions on document collections.</li>
 * </ul>
 * <p>
 * It is highly recommended <b>not</b> to remove stopwords (very common terms: the, of, and, to, a, in, for, is, on, that, etc) with this similarity.
 * <p>
 * For more information see: <a href="http://dx.doi.org/10.1007/s10791-013-9225-4">A nonparametric term weighting method for information retrieval based on measuring the divergence from independence</a>
 *
 * @lucene.experimental
 * @see org.apache.lucene.search.similarities.IndependenceStandardized
 * @see org.apache.lucene.search.similarities.IndependenceSaturated
 * @see org.apache.lucene.search.similarities.IndependenceChiSquared
 */
type dfiSimilarityImpl struct {
	*similarityBaseImpl
	independence Independence
}

func NewDFISimilarity(independence Independence) *dfiSimilarityImpl {
	if independence == nil {
		panic("nil  parameters not allowed.")
	}
	ans := &dfiSimilarityImpl{independence: independence}
	ans.similarityBaseImpl = NewSimilarityBase(ans)
	return ans
}

func (dfi *dfiSimilarityImpl) score(stats Stats, freq, docLen float32) float32 {
	expected := (float32(stats.TotalTermFreq()) + 1) * docLen / (float32(stats.NumberOfFieldTokens()) + 1)

	// if the observed frequency is less than or equal to the expected value, then return zero.
	if freq <= expected {
		return 0
	}

	measure := dfi.independence.Score(freq, expected)
	// idf := float32(math.Log2(float64(stats.NumberOfDocuments())/float64(freq+1))) + 1.0
	return stats.TotalBoost() * float32(math.Log2(float64(measure)+1)) // * idf
}

func (dfi *dfiSimilarityImpl) String() string {
	return fmt.Sprintf("DFI(%s)", dfi.independence.String())
}
