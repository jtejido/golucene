package similarities

import (
	"fmt"
	"math"
)

var _ Similarity = (*ModBM25Similarity)(nil)

/**
 * ModBM25 is a modified version of BM25 that ensures negative IDF don't violate Term-Frequency, Length Normalization and
 * TF-LENGTH Constraints by using Robertson-Sparck Idf.
 *
 * The implementation is based on the paper by Fang Et al.,
 * @see http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.59.1189&rep=rep1&type=pdf
 *
 * @lucene.experimental(jtejido)
 */
type ModBM25Similarity struct {
	baseBM25Similarity
}

func NewDefaultModBM25Similarity() *ModBM25Similarity {
	ans := NewModBM25Similarity(DEFAULT_K1, DEFAULT_B, true)
	return ans
}

func NewModBM25Similarity(k1, b float32, discountOverlaps bool) *ModBM25Similarity {
	ans := new(ModBM25Similarity)
	ans.k1 = k1
	ans.b = b
	ans.discountOverlaps = discountOverlaps
	ans.spi = ans
	norm_table = ans.buildNormTable()
	return ans
}

/**
 * We'll use pivoted normalized Idf as BM25's Idf.
 */
func (bm25 *ModBM25Similarity) idf(docFreq, numDocs int64) float32 {
	return float32(math.Log(float64(numDocs+1) / float64(docFreq)))
}

func (bm25 *ModBM25Similarity) score(freq, norm float32) float32 {
	num := freq * (bm25.k1 + 1)
	denom := freq + bm25.k1*(1-bm25.b+bm25.b*norm)

	return num / denom
}

func (bm25 *ModBM25Similarity) String() string {
	return fmt.Sprintf("ModBM25(k1=%v, b=%v)", bm25.k1, bm25.b)
}
