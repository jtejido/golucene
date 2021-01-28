package similarities

import (
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/util"
	"math"
)

/** Cache of decoded bytes. */
var NORM_TABLE []float32

type DefaultSimilarity struct {
	*TFIDFSimilarity
	discountOverlaps bool
}

func NewDefaultSimilarity() *DefaultSimilarity {
	ans := &DefaultSimilarity{discountOverlaps: true}
	ans.TFIDFSimilarity = newTFIDFSimilarity(ans)
	NORM_TABLE = ans.buildNormTable()
	return ans
}

func (ds *DefaultSimilarity) buildNormTable() []float32 {
	table := make([]float32, 256)
	for i, _ := range table {
		table[i] = util.Byte315ToFloat(byte(i))
	}
	return table
}

func (ds *DefaultSimilarity) Coord(overlap, maxOverlap int) float32 {
	return float32(overlap) / float32(maxOverlap)
}

func (ds *DefaultSimilarity) QueryNorm(sumOfSquaredWeights float32) float32 {
	return float32(1.0 / math.Sqrt(float64(sumOfSquaredWeights)))
}

/*
Encodes a normalization factor for storage in an index.

The encoding uses a three-bit mantissa, a five-bit exponent, and the
zero-exponent point at 15, thus representing values from around
7x10^9 to 2x10^-9 with about one significant decimal digit of
accuracy. Zero is also represented. Negative numbers are rounded up
to zero. Values too large to represent are rounded down to the
largest representable value. Positive values too small to represent
are rounded up to the smallest positive representable value.
*/
func (ds *DefaultSimilarity) encodeNormValue(f float32) int64 {
	return int64(util.FloatToByte315(f))
}

func (ds *DefaultSimilarity) decodeNormValue(norm int64) float32 {
	return NORM_TABLE[int(norm&0xff)] // & 0xFF maps negative bytes to positive above 127
}

/*
Implemented as state.boost() * lengthNorm(numTerms), where numTerms
is FieldInvertState.length() if setDiscountOverlaps() is false, else
it's FieldInvertState.length() - FieldInvertState.numOverlap().
*/
func (ds *DefaultSimilarity) lengthNorm(state *index.FieldInvertState) float32 {
	var numTerms int
	if ds.discountOverlaps {
		numTerms = state.Length() - state.NumOverlap()
	} else {
		numTerms = state.Length()
	}
	return state.Boost() * float32(1.0/math.Sqrt(float64(numTerms)))
}

func (ds *DefaultSimilarity) tf(freq float32) float32 {
	return float32(math.Sqrt(float64(freq)))
}

func (ds *DefaultSimilarity) idf(docFreq int64, numDocs int64) float32 {
	return float32(math.Log(float64(numDocs)/float64(docFreq+1))) + 1.0
}

func (ds *DefaultSimilarity) String() string {
	return "DefaultSImilarity"
}
