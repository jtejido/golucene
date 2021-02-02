package search

import (
	"fmt"
	"github.com/jtejido/golucene/core/index"
	"log"
	"math"
)

/* Define service that can be overrided */
type IndexSearcherSPI interface {
	CreateNormalizedWeight(Query) (Weight, error)
	Rewrite(Query) (Query, error)
	WrapFilter(Query, Filter) Query
	SearchLWC([]*index.AtomicReaderContext, Weight, Collector) error
}

// IndexSearcher
type IndexSearcher struct {
	spi           IndexSearcherSPI
	reader        index.IndexReader
	readerContext index.IndexReaderContext
	leafContexts  []*index.AtomicReaderContext
	similarity    Similarity
}

func NewIndexSearcher(r index.IndexReader) *IndexSearcher {
	// log.Print("Initializing IndexSearcher from IndexReader: ", r)
	return NewIndexSearcherFromContext(r.Context())
}

func NewIndexSearcherFromContext(context index.IndexReaderContext) *IndexSearcher {
	// assert2(context.isTopLevel, "IndexSearcher's ReaderContext must be topLevel for reader %v", context.reader())
	defaultSimilarity := index.DefaultSimilarity().(Similarity)
	ss := &IndexSearcher{nil, context.Reader(), context, context.Leaves(), defaultSimilarity}
	ss.spi = ss
	return ss
}

/* Expert: set the similarity implementation used by this IndexSearcher. */
func (ss *IndexSearcher) SetSimilarity(similarity Similarity) {
	ss.similarity = similarity
}

func (ss *IndexSearcher) SearchTop(q Query, n int) (topDocs TopDocs, err error) {
	return ss.Search(q, nil, n)
}

func (ss *IndexSearcher) Search(q Query, f Filter, n int) (topDocs TopDocs, err error) {
	w, err := ss.spi.CreateNormalizedWeight(ss.spi.WrapFilter(q, f))
	if err != nil {
		return TopDocs{}, err
	}
	return ss.searchWSI(w, nil, n), nil
}

/** Expert: Low-level search implementation.  Finds the top <code>n</code>
 * hits for <code>query</code>, applying <code>filter</code> if non-null.
 *
 * <p>Applications should usually call {@link IndexSearcher#search(Query,int)} or
 * {@link IndexSearcher#search(Query,Filter,int)} instead.
 * @throws BooleanQuery.TooManyClauses If a query would exceed
 *         {@link BooleanQuery#getMaxClauseCount()} clauses.
 */
func (ss *IndexSearcher) searchWSI(w Weight, after *ScoreDoc, nDocs int) TopDocs {
	// TODO support concurrent search
	return ss.searchLWSI(ss.leafContexts, w, after, nDocs)
}

/** Expert: Low-level search implementation.  Finds the top <code>n</code>
 * hits for <code>query</code>.
 *
 * <p>Applications should usually call {@link IndexSearcher#search(Query,int)} or
 * {@link IndexSearcher#search(Query,Filter,int)} instead.
 * @throws BooleanQuery.TooManyClauses If a query would exceed
 *         {@link BooleanQuery#getMaxClauseCount()} clauses.
 */
func (ss *IndexSearcher) searchLWSI(leaves []*index.AtomicReaderContext,
	w Weight, after *ScoreDoc, nDocs int) TopDocs {
	// single thread
	limit := ss.reader.MaxDoc()
	if limit == 0 {
		limit = 1
	}
	if nDocs > limit {
		nDocs = limit
	}
	collector := NewTopScoreDocCollector(nDocs, after, !w.IsScoresDocsOutOfOrder())
	ss.spi.SearchLWC(leaves, w, collector)
	return collector.TopDocs()
}

func (ss *IndexSearcher) SearchLWC(leaves []*index.AtomicReaderContext, w Weight, c Collector) (err error) {
	// TODO: should we make this
	// threaded...?  the Collector could be sync'd?
	// always use single thread:
	for _, ctx := range leaves { // search each subreader
		// TODO catch CollectionTerminatedException
		c.SetNextReader(ctx)

		scorer, err := w.BulkScorer(ctx, !c.AcceptsDocsOutOfOrder(),
			ctx.Reader().(index.AtomicReader).LiveDocs())
		if err != nil {
			return err
		}
		if scorer != nil {
			err = scorer.ScoreAndCollect(c)
		} // TODO catch CollectionTerminatedException
	}
	return
}

func (ss *IndexSearcher) WrapFilter(q Query, f Filter) Query {
	if f == nil {
		return q
	}
	panic("FilteredQuery not supported yet")
}

/*
Returns an Explanation that describes how doc scored against query.

This is intended to be used in developing Similiarity implemenations, and, for
good performance, should not be displayed with every hit. Computing an
explanation is as expensive as executing the query over the entire index.
*/
func (ss *IndexSearcher) Explain(query Query, doc int) (exp Explanation, err error) {
	w, err := ss.spi.CreateNormalizedWeight(query)
	if err == nil {
		return ss.explain(w, doc)
	}
	return
}

/*
Expert: low-level implementation method
Returns an Explanation that describes how doc scored against weight.

This is intended to be used in developing Similarity implementations, and, for
good performance, should not be displayed with every hit. Computing an
explanation is as expensive as executing the query over the entire index.

Applications should call explain(Query, int).
*/
func (ss *IndexSearcher) explain(weight Weight, doc int) (exp Explanation, err error) {
	n := index.SubIndex(doc, ss.leafContexts)
	ctx := ss.leafContexts[n]
	deBasedDoc := doc - ctx.DocBase
	return weight.Explain(ctx, deBasedDoc)
}

func (ss *IndexSearcher) CreateNormalizedWeight(q Query) (w Weight, err error) {
	q, err = ss.spi.Rewrite(q)
	if err != nil {
		return nil, err
	}
	log.Printf("After rewrite: %v", q)
	w, err = q.CreateWeight(ss)
	if err != nil {
		return nil, err
	}
	v := w.ValueForNormalization()
	norm := ss.similarity.QueryNorm(v)
	if math.IsInf(float64(norm), 1) || math.IsNaN(float64(norm)) {
		norm = 1.0
	}
	w.Normalize(norm, 1.0)
	return w, nil
}

func (ss *IndexSearcher) Rewrite(q Query) (Query, error) {
	log.Printf("Rewriting '%v'...", q)
	after := q.Rewrite(ss.reader)
	for after != q {
		q = after
		after = q.Rewrite(ss.reader)
	}
	return q, nil
}

// Returns this searhcers the top-level IndexReaderContext
func (ss *IndexSearcher) TopReaderContext() index.IndexReaderContext {
	return ss.readerContext
}

func (ss *IndexSearcher) String() string {
	return fmt.Sprintf("IndexSearcher(%v)", ss.reader)
}

func (ss *IndexSearcher) TermStatistics(term *index.Term, context *index.TermContext) TermStatistics {
	return NewTermStatistics(term.Bytes, int64(context.DocFreq), context.TotalTermFreq)
}

func (ss *IndexSearcher) CollectionStatistics(field string) CollectionStatistics {
	terms := index.GetMultiTerms(ss.reader, field)
	if terms == nil {
		return NewCollectionStatistics(field, int64(ss.reader.MaxDoc()), 0, 0, 0)
	}
	return NewCollectionStatistics(field, int64(ss.reader.MaxDoc()), int64(terms.DocCount()), terms.SumTotalTermFreq(), terms.SumDocFreq())
}
