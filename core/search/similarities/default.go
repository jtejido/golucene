package similarities

import (
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/util"
	"math"
	"sync"
)

var _ Similarity = (*DefaultSimilarity)(nil)

var norm_table []float32
var once sync.Once

/**
 * Expert: Default scoring implementation which {@link #encodeNormValue(float)
 * encodes} norm values as a single byte before being stored. At search time,
 * the norm byte value is read from the index
 * {@link org.apache.lucene.store.Directory directory} and
 * {@link #decodeNormValue(long) decoded} back to a float <i>norm</i> value.
 * This encoding/decoding, while reducing index size, comes with the price of
 * precision loss - it is not guaranteed that <i>decode(encode(x)) = x</i>. For
 * instance, <i>decode(encode(0.89)) = 0.75</i>.
 * <p>
 * Compression of norm values to a single byte saves memory at search time,
 * because once a field is referenced at search time, its norms - for all
 * documents - are maintained in memory.
 * <p>
 * The rationale supporting such lossy compression of norm values is that given
 * the difficulty (and inaccuracy) of users to express their true information
 * need by a query, only big differences matter. <br>
 * &nbsp;<br>
 * Last, note that search time is too late to modify this <i>norm</i> part of
 * scoring, e.g. by using a different {@link Similarity} for search.
 */
type DefaultSimilarity struct {
	TFIDFSimilarityImpl
	discountOverlaps bool
}

func NewDefaultSimilarity() *DefaultSimilarity {
	ans := new(DefaultSimilarity)
	ans.owner = ans

	once.Do(func() {
		norm_table = ans.buildNormTable()
	})
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
	return norm_table[int(norm&0xff)] // & 0xFF maps negative bytes to positive above 127
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

func (ds *DefaultSimilarity) sloppyFreq(distance int) float32 {
	return 1.0 / (float32(distance) + 1)
}

func (ds *DefaultSimilarity) scorePayload(doc, start, end int, payload *util.BytesRef) float32 {
	return 1
}

func (ds *DefaultSimilarity) String() string {
	return "DefaultSImilarity"
}
