package similarities

import (
	"github.com/jtejido/golucene/core/search"
)

type IBasicStats interface {
	search.SimWeight
	NumberOfDocuments() int64
	SetNumberOfDocuments(numberOfDocuments int64)
	NumberOfFieldTokens() int64
	SetNumberOfFieldTokens(numberOfFieldTokens int64)
	AvgFieldLength() float32
	SetAvgFieldLength(avgFieldLength float32)
	DocFreq() int64
	SetDocFreq(docFreq int64)
	TotalTermFreq() int64
	SetTotalTermFreq(totalTermFreq int64)
	CollectionProbability() float32
	TotalBoost() float32
}

type BasicStats struct {
	spi   ILMStats
	field string
	/** The number of documents. */
	numberOfDocuments int64
	/** The total number of tokens in the field. */
	numberOfFieldTokens int64
	/** The average field length. */
	avgFieldLength float32
	/** The document frequency. */
	docFreq int64
	/** The total number of occurrences of this term across all documents. */
	totalTermFreq int64
	/** Query's inner boost. */
	queryBoost float32
	/** Any outer query's boost. */
	topLevelBoost float32
	/** For most Similarities, the immediate and the top level query boosts are
	 * not handled differently. Hence, this field is just the product of the
	 * other two. */
	totalBoost float32
}

func newBasicStats(spi ILMStats, field string, queryBoost float32) *BasicStats {
	return &BasicStats{spi: spi, field: field, queryBoost: queryBoost, totalBoost: queryBoost}
}

func (bs *BasicStats) Field() string {
	return bs.field
}

// ------------------------- Getter/setter methods -------------------------

/** Returns the number of documents. */
func (bs *BasicStats) NumberOfDocuments() int64 {
	return bs.numberOfDocuments
}

/** Sets the number of documents. */
func (bs *BasicStats) SetNumberOfDocuments(numberOfDocuments int64) {
	bs.numberOfDocuments = numberOfDocuments
}

/**
 * Returns the total number of tokens in the field.
 * @see Terms#getSumTotalTermFreq()
 */
func (bs *BasicStats) NumberOfFieldTokens() int64 {
	return bs.numberOfFieldTokens
}

/**
 * Sets the total number of tokens in the field.
 * @see Terms#getSumTotalTermFreq()
 */
func (bs *BasicStats) SetNumberOfFieldTokens(numberOfFieldTokens int64) {
	bs.numberOfFieldTokens = numberOfFieldTokens
}

/** Returns the average field length. */
func (bs *BasicStats) AvgFieldLength() float32 {
	return bs.avgFieldLength
}

/** Sets the average field length. */
func (bs *BasicStats) SetAvgFieldLength(avgFieldLength float32) {
	bs.avgFieldLength = avgFieldLength
}

/** Returns the document frequency. */
func (bs *BasicStats) DocFreq() int64 {
	return bs.docFreq
}

/** Sets the document frequency. */
func (bs *BasicStats) SetDocFreq(docFreq int64) {
	bs.docFreq = docFreq
}

/** Returns the total number of occurrences of this term across all documents. */
func (bs *BasicStats) TotalTermFreq() int64 {
	return bs.totalTermFreq
}

/** Sets the total number of occurrences of this term across all documents. */
func (bs *BasicStats) SetTotalTermFreq(totalTermFreq int64) {
	bs.totalTermFreq = totalTermFreq
}

func (bs *BasicStats) CollectionProbability() float32 {
	return bs.spi.CollectionProbability()
}

// -------------------------- Boost-related stuff --------------------------

/** The square of the raw normalization value.
 * @see #rawNormalizationValue() */
func (bs *BasicStats) ValueForNormalization() float32 {
	rawValue := bs.rawNormalizationValue()
	return rawValue * rawValue
}

/** Computes the raw normalization value. This basic implementation returns
 * the query boost. Subclasses may override this method to include other
 * factors (such as idf), or to save the value for inclusion in
 * {@link #normalize(float, float)}, etc.
 */
func (bs *BasicStats) rawNormalizationValue() float32 {
	return bs.queryBoost
}

/** No normalization is done. {@code topLevelBoost} is saved in the object,
 * however. */
func (bs *BasicStats) Normalize(norm float32, topLevelBoost float32) {
	bs.topLevelBoost = topLevelBoost
	bs.totalBoost = bs.queryBoost * topLevelBoost
}

/** Returns the total boost. */
func (bs *BasicStats) TotalBoost() float32 {
	return bs.totalBoost
}
