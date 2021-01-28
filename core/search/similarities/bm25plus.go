package similarities

import (
	"fmt"
	"math"
)

const (
	DEFAULT_BM25PLUS_D float32 = 1
)

/**
 * BM25 is a class for ranking documents against a query where we made use of a delta(δ) value of 1,
 * which modifies BM25 to account for an issue against penalizing long documents and allowing shorter ones to dominate.
 * The delta values assures BM25 to be lower-bounded.
 * @see http://sifaka.cs.uiuc.edu/~ylv2/pub/cikm11-lowerbound.pdf
 *
 * Some modifications have been made to allow for non-negative scoring as suggested here.
 * @see https://doc.rero.ch/record/16754/files/Dolamic_Ljiljana_-_When_Stopword_Lists_Make_the_Difference_20091218.pdf
 *
 * We made use of a delta(δ) value of 1, which modifies BM25 to account for an issue against
 * penalizing long documents and allowing shorter ones to dominate. The delta values assures BM25
 * to be lower-bounded. (This makes this class BM25+)
 * @see http://sifaka.cs.uiuc.edu/~ylv2/pub/cikm11-lowerbound.pdf
 *
 */
type BM25PlusSimilarity struct {
	*ProbabilitySimilarity
	d float32
}

func NewDefaultBM25PlusSimilarity() *BM25PlusSimilarity {
	ans := NewBM25PlusSimilarity(DEFAULT_K1, DEFAULT_B, DEFAULT_BM25PLUS_D, true)
	return ans
}

func NewBM25PlusSimilarity(k1, b, d float32, discountOverlaps bool) *BM25PlusSimilarity {
	ans := new(BM25PlusSimilarity)
	ans.ProbabilitySimilarity = newProbabilitySimilarity(ans, k1, b, discountOverlaps)
	ans.d = d
	return ans
}

func (bmp *BM25PlusSimilarity) idf(docFreq, numDocs int64) float32 {
	return float32(math.Log(1. + (float64(numDocs)-float64(docFreq)+0.5)/(float64(docFreq)+0.5)))
}

func (bmp *BM25PlusSimilarity) score(freq, norm float32) float32 {
	num := freq * (bmp.k1 + 1)
	denom := freq + bmp.k1*(1-bmp.b+bmp.b*norm)

	return ((num / denom) + bmp.d)
}

func (bmp *BM25PlusSimilarity) String() string {
	return fmt.Sprintf("BM25+(k1=%v, b=%v, d=%v)", bmp.k1, bmp.b, bmp.d)
}
