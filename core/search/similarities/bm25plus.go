package similarities

import (
	"fmt"
)

var _ Similarity = (*BM25PlusSimilarity)(nil)

/**
 * BM25 is a class for ranking documents against a query where we made use of a delta(δ) value of 1,
 * which modifies BM25 to account for an issue against penalizing long documents and allowing shorter ones to dominate.
 * The delta values assures BM25 to be lower-bounded.
 * @see http://sifaka.cs.uiuc.edu/~ylv2/pub/cikm11-lowerbound.pdf
 *
 * We made use of a delta(δ) value of 1, which modifies BM25 to account for an issue against
 * penalizing long documents and allowing shorter ones to dominate. The delta values assures BM25
 * to be lower-bounded. (This makes this class BM25+)
 * @see http://sifaka.cs.uiuc.edu/~ylv2/pub/cikm11-lowerbound.pdf
 *
 */
type BM25PlusSimilarity struct {
	BM25Similarity
}

func NewDefaultBM25PlusSimilarity() *BM25PlusSimilarity {
	return NewBM25PlusSimilarity(DEFAULT_K1, DEFAULT_B, true)
}

func NewBM25PlusSimilarity(k1, b float32, discountOverlaps bool) *BM25PlusSimilarity {
	ans := new(BM25PlusSimilarity)
	ans.k1 = k1
	ans.b = b
	ans.discountOverlaps = discountOverlaps
	ans.spi = ans
	norm_table = ans.buildNormTable()
	return ans
}

func (bmp *BM25PlusSimilarity) score(freq, norm float32) float32 {
	return bmp.BM25Similarity.score(freq, norm) + 1
}

func (bmp *BM25PlusSimilarity) String() string {
	return fmt.Sprintf("BM25+(k1=%v, b=%v)", bmp.k1, bmp.b)
}
