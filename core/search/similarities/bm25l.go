package similarities

import (
	"fmt"
	"math"
)

const (
	DEFAULT_D float32 = .5
)

/**
 * BM25L is a work of Lv and Zhai to rewrite BM25 due to Singhal et al's observation for having it penalized
 * longer documents.
 *
 * When Documents Are Very Long, BM25 Fails! (Lv and Zhai).
 * @see http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.600.16&rep=rep1&type=pdf
 *
 * A. Singhal, C. Buckley, and M. Mitra. Pivoted document length normalization. In SIGIR ’96, pages 21–29, 1996.
 */
type BM25LSimilarity struct {
	*ProbabilitySimilarity
	d float32
}

func NewDefaultBM25LSimilarity() *BM25LSimilarity {
	ans := NewBM25LSimilarity(DEFAULT_K1, DEFAULT_B, DEFAULT_D, true)
	return ans
}

func NewBM25LSimilarity(k1, b, d float32, discountOverlaps bool) *BM25LSimilarity {

	ans := new(BM25LSimilarity)
	ans.ProbabilitySimilarity = newProbabilitySimilarity(ans, k1, b, discountOverlaps)
	ans.d = d
	return ans

}

func (bm25l *BM25LSimilarity) idf(docFreq, numDocs int64) float32 {
	return float32(math.Log((float64(numDocs) + 1) / (float64(docFreq) + 0.5)))
}

func (bm25l *BM25LSimilarity) score(freq, norm float32) float32 {
	tfNormalized := freq / (1 - bm25l.b + bm25l.b*norm)

	num := (bm25l.k1 + 1) * (tfNormalized + bm25l.d)
	denom := bm25l.k1 + (tfNormalized + bm25l.d)

	return num / denom
}

func (bm25l *BM25LSimilarity) String() string {
	return fmt.Sprintf("BM25L(k1=%v, b=%v, d=%v)", bm25l.k1, bm25l.b, bm25l.d)
}
