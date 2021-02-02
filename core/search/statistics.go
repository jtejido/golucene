package search

type TermStatistics struct {
	Term                   []byte
	DocFreq, TotalTermFreq int64
}

func NewTermStatistics(term []byte, docFreq, totalTermFreq int64) TermStatistics {
	// assert docFreq >= 0;
	// assert totalTermFreq == -1 || totalTermFreq >= docFreq; // #positions must be >= #postings
	return TermStatistics{term, docFreq, totalTermFreq}
}

type CollectionStatistics struct {
	field                                          string
	maxDoc, docCount, sumTotalTermFreq, sumDocFreq int64
}

func NewCollectionStatistics(field string, maxDoc, docCount, sumTotalTermFreq, sumDocFreq int64) CollectionStatistics {
	// assert maxDoc >= 0;
	// assert docCount >= -1 && docCount <= maxDoc; // #docs with field must be <= #docs
	// assert sumDocFreq == -1 || sumDocFreq >= docCount; // #postings must be >= #docs with field
	// assert sumTotalTermFreq == -1 || sumTotalTermFreq >= sumDocFreq; // #positions must be >= #postings
	return CollectionStatistics{field, maxDoc, docCount, sumTotalTermFreq, sumDocFreq}
}

func (cs CollectionStatistics) Field() string {
	return cs.field
}

func (cs CollectionStatistics) MaxDoc() int64 {
	return cs.maxDoc
}

func (cs CollectionStatistics) DocCount() int64 {
	return cs.docCount
}

func (cs CollectionStatistics) SumTotalTermFreq() int64 {
	return cs.sumTotalTermFreq
}

func (cs CollectionStatistics) SumDocFreq() int64 {
	return cs.sumDocFreq
}
