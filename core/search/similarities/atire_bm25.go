package similarities

import (
	"fmt"
	"math"
)

/**
 * ATIRE BM25 is a class that uses Robertson-Walker IDF instead of the original Robertson-Sparck IDF.
 *
 * Towards an Efficient and Effective Search Engine (Trotman, Jia, Crane).
 * SIGIR 2012 Workshop on Open Source Information Retrieval.
 * @see http://opensearchlab.otago.ac.nz/paper_4.pdf
 */
type AtireBM25Similarity struct {
	*ProbabilitySimilarity
}

func NewDefaultAtireBM25Similarity() *AtireBM25Similarity {
	ans := NewAtireBM25Similarity(DEFAULT_K1, DEFAULT_B, DEFAULT_D, true)
	return ans
}

func NewAtireBM25Similarity(k1, b, d float32, discountOverlaps bool) *AtireBM25Similarity {
	ans := new(AtireBM25Similarity)
	ans.ProbabilitySimilarity = newProbabilitySimilarity(ans, k1, b, discountOverlaps)
	return ans
}

func (abm *AtireBM25Similarity) idf(docFreq, numDocs int64) float32 {
	return float32(math.Log(float64(numDocs) / float64(docFreq)))
}

func (abm *AtireBM25Similarity) score(freq, norm float32) float32 {
	num := freq * (abm.k1 + 1)
	denom := freq + abm.k1*(1-abm.b+abm.b*norm)

	return num / denom
}

func (abm *AtireBM25Similarity) String() string {
	return fmt.Sprintf("ATIRE BM25(k1=%v, b=%v)", abm.k1, abm.b)
}
