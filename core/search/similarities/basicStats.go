package similarities

var _ Stats = (*basicStatsImpl)(nil)

type basicStatsImpl struct {
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

func NewBasicStats(field string, queryBoost float32) *basicStatsImpl {
	return &basicStatsImpl{field: field, queryBoost: queryBoost, totalBoost: queryBoost}
}

func (bs *basicStatsImpl) Field() string {
	return bs.field
}

// ------------------------- Getter/setter methods -------------------------

/** Returns the number of documents. */
func (bs *basicStatsImpl) NumberOfDocuments() int64 {
	return bs.numberOfDocuments
}

/** Sets the number of documents. */
func (bs *basicStatsImpl) SetNumberOfDocuments(numberOfDocuments int64) {
	bs.numberOfDocuments = numberOfDocuments
}

/**
 * Returns the total number of tokens in the field.
 * @see Terms#getSumTotalTermFreq()
 */
func (bs *basicStatsImpl) NumberOfFieldTokens() int64 {
	return bs.numberOfFieldTokens
}

/**
 * Sets the total number of tokens in the field.
 * @see Terms#getSumTotalTermFreq()
 */
func (bs *basicStatsImpl) SetNumberOfFieldTokens(numberOfFieldTokens int64) {
	bs.numberOfFieldTokens = numberOfFieldTokens
}

/** Returns the average field length. */
func (bs *basicStatsImpl) AvgFieldLength() float32 {
	return bs.avgFieldLength
}

/** Sets the average field length. */
func (bs *basicStatsImpl) SetAvgFieldLength(avgFieldLength float32) {
	bs.avgFieldLength = avgFieldLength
}

/** Returns the document frequency. */
func (bs *basicStatsImpl) DocFreq() int64 {
	return bs.docFreq
}

/** Sets the document frequency. */
func (bs *basicStatsImpl) SetDocFreq(docFreq int64) {
	bs.docFreq = docFreq
}

/** Returns the total number of occurrences of this term across all documents. */
func (bs *basicStatsImpl) TotalTermFreq() int64 {
	return bs.totalTermFreq
}

/** Sets the total number of occurrences of this term across all documents. */
func (bs *basicStatsImpl) SetTotalTermFreq(totalTermFreq int64) {
	bs.totalTermFreq = totalTermFreq
}

// -------------------------- Boost-related stuff --------------------------

/** The square of the raw normalization value.
 * @see #rawNormalizationValue() */
func (bs *basicStatsImpl) ValueForNormalization() float32 {
	rawValue := bs.rawNormalizationValue()
	return rawValue * rawValue
}

/** Computes the raw normalization value. This basic implementation returns
 * the query boost. Subclasses may override this method to include other
 * factors (such as idf), or to save the value for inclusion in
 * {@link #normalize(float, float)}, etc.
 */
func (bs *basicStatsImpl) rawNormalizationValue() float32 {
	return bs.queryBoost
}

/** No normalization is done. {@code topLevelBoost} is saved in the object,
 * however. */
func (bs *basicStatsImpl) Normalize(norm float32, topLevelBoost float32) {
	bs.topLevelBoost = topLevelBoost
	bs.totalBoost = bs.queryBoost * topLevelBoost
}

/** Returns the total boost. */
func (bs *basicStatsImpl) TotalBoost() float32 {
	return bs.totalBoost
}
