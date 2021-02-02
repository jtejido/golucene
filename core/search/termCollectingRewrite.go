package search

import (
	"fmt"
	"github.com/jtejido/golucene/core/index"
	. "github.com/jtejido/golucene/core/index/model"
	"github.com/jtejido/golucene/core/util"
	"sort"
)

type TermCollectingRewrite interface {
	RewriteMethod
	/** Return a suitable top-level Query for holding all expanded terms. */
	TopLevelQuery() Query
	/** Add a MultiTermQuery term to the top-level query */
	AddClause(topLevel Query, term *index.Term, docCount int, boost float32)
	AddClauseWithContext(topLevel Query, term *index.Term, docCount int, boost float32, states *index.TermContext)
}

type TermCollectingRewriteSPI interface {
	TopLevelQuery() Query
	AddClauseWithContext(topLevel Query, term *index.Term, docCount int, boost float32, states *index.TermContext)
	TermsEnum(query MultiTermQuery, terms Terms, atts *util.AttributeSource) (TermsEnum, error)
}

type AbstractTermCollectingRewrite struct {
	spi TermCollectingRewriteSPI
}

func newAbstractTermCollectingRewrite(spi TermCollectingRewriteSPI) *AbstractTermCollectingRewrite {
	return &AbstractTermCollectingRewrite{spi}
}

func (a *AbstractTermCollectingRewrite) AddClause(topLevel Query, term *index.Term, docCount int, boost float32) {
	a.spi.AddClauseWithContext(topLevel, term, docCount, boost, nil)
}

func (a *AbstractTermCollectingRewrite) CollectTerms(reader index.IndexReader, query MultiTermQuery, collector TermCollector) error {
	topReaderContext := reader.Context()
	var lastTermComp sort.Interface
	for _, context := range topReaderContext.Leaves() {
		fields := context.Reader().(index.AtomicReader).Fields()
		if fields == nil {
			// reader has no fields
			continue
		}

		terms := fields.Terms(query.Field())
		if terms == nil {
			// field does not exist
			continue
		}

		termsEnum, err := a.spi.TermsEnum(query, terms, collector.Attributes())
		if err != nil {
			return err
		}
		assert(termsEnum != nil)

		if termsEnum == EMPTY_TERMS_ENUM {
			continue
		}

		// Check comparator compatibility:
		newTermComp := termsEnum.Comparator()
		if lastTermComp != nil && newTermComp != nil && newTermComp != lastTermComp {
			return fmt.Errorf("term comparator should not change between segments: %v != %v", lastTermComp, newTermComp)
		}
		lastTermComp = newTermComp
		collector.SetReaderContext(topReaderContext, context)
		collector.SetNextEnum(termsEnum)

		for {
			bytes, err := termsEnum.Next()
			if err != nil {
				return err
			}
			if bytes == nil {
				break
			}
			if !collector.Collect(util.NewBytesRefFrom(bytes)) {
				return nil
			}
		}
	}

	return nil
}

type TermCollector interface {
	Collect(*util.BytesRef) bool
	SetNextEnum(termsEnum TermsEnum)
	SetReaderContext(topReaderContext index.IndexReaderContext, readerContext *index.AtomicReaderContext)
	Attributes() *util.AttributeSource
}

type TermCollectorSPI interface {
	Collect(*util.BytesRef) bool
	SetNextEnum(termsEnum TermsEnum)
}

type AbstractTermCollector struct {
	spi              TermCollectorSPI
	readerContext    *index.AtomicReaderContext
	topReaderContext index.IndexReaderContext
	attributes       *util.AttributeSource
}

func newAbstractTermCollector(spi TermCollectorSPI) *AbstractTermCollector {
	return &AbstractTermCollector{
		spi: spi,
	}
}

func (c *AbstractTermCollector) SetReaderContext(topReaderContext index.IndexReaderContext, readerContext *index.AtomicReaderContext) {
	c.readerContext = readerContext
	c.topReaderContext = topReaderContext
}

func (c *AbstractTermCollector) SetNextEnum(termsEnum TermsEnum) {
	c.spi.SetNextEnum(termsEnum)
}

func (c *AbstractTermCollector) Collect(bytes *util.BytesRef) bool {
	return c.spi.Collect(bytes)
}
