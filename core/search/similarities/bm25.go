package similarities

import (
	"fmt"
	"math"
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
	*ProbabilitySimilarity
}

func NewDefaultBM25Similarity() *BM25Similarity {
	ans := NewBM25Similarity(DEFAULT_K1, DEFAULT_B, true)
	return ans
}

func NewBM25Similarity(k1, b float32, discountOverlaps bool) *BM25Similarity {
	ans := new(BM25Similarity)
	ans.ProbabilitySimilarity = newProbabilitySimilarity(ans, k1, b, discountOverlaps)
	return ans
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
