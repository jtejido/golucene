package search

import (
	"bytes"
	"fmt"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/index/model"
	"github.com/jtejido/golucene/core/util"
	"sort"
	"strconv"
)

type PhraseQuery struct {
	*AbstractQuery
	field       string
	terms       []*index.Term
	positions   []int32
	maxPosition int32
	slop        int
}

func NewPhraseQuery() *PhraseQuery {
	ans := new(PhraseQuery)
	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}

func (q *PhraseQuery) Add(term *index.Term) {
	position := int32(0)
	if len(q.positions) > 0 {
		position = q.positions[len(q.positions)-1] + 1
	}

	q.AddTermWithPosition(term, position)
}

func (q *PhraseQuery) AddTermWithPosition(term *index.Term, position int32) {
	if len(q.terms) == 0 {
		q.field = term.Field
	} else if term.Field != q.field {
		panic(fmt.Sprintf("All phrase terms must be in the same field: %s", term.String()))
	}

	q.terms = append(q.terms, term)
	q.positions = append(q.positions, position)
	if position > q.maxPosition {
		q.maxPosition = position
	}
}

type PhraseWeight struct {
	*WeightImpl
	owner      *PhraseQuery
	similarity Similarity
	stats      SimWeight
	states     []*index.TermContext
}

func newPhraseWeight(owner *PhraseQuery, searcher *IndexSearcher) (w *PhraseWeight, err error) {
	w = &PhraseWeight{
		owner:      owner,
		similarity: searcher.similarity,
		states:     make([]*index.TermContext, len(owner.terms)),
	}

	context := searcher.TopReaderContext()
	termStats := make([]TermStatistics, len(owner.terms))
	for i := 0; i < len(owner.terms); i++ {
		term := owner.terms[i]
		w.states[i], err = index.NewTermContextFromTerm(context, term)
		if err != nil {
			return nil, err
		}
		termStats[i] = searcher.TermStatistics(term, w.states[i])
	}
	w.stats = w.similarity.ComputeWeight(owner.Boost(), searcher.CollectionStatistics(owner.field), termStats...)
	w.WeightImpl = newWeightImpl(w)
	return w, nil
}

func (w *PhraseWeight) Explain(context *index.AtomicReaderContext, doc int) (Explanation, error) {
	panic("not implemented yet")
}

func (w *PhraseWeight) IsScoresDocsOutOfOrder() bool {
	return false
}

func (w *PhraseWeight) ValueForNormalization() (sum float32) {
	return w.stats.ValueForNormalization()
}

func (w *PhraseWeight) Normalize(norm, topLevelBoost float32) {
	w.stats.Normalize(norm, topLevelBoost)
}

func (w *PhraseWeight) Scorer(context *index.AtomicReaderContext, acceptDocs util.Bits) (scorer Scorer, err error) {
	assert(len(w.owner.terms) != 0)
	reader := context.Reader()
	liveDocs := acceptDocs
	postingsFreqs := make([]*PostingsAndFreq, len(w.owner.terms))

	fieldTerms := reader.(index.AtomicReader).Terms(w.owner.field)
	if fieldTerms == nil {
		return nil, nil
	}

	// Reuse single TermsEnum below:
	te := fieldTerms.Iterator(nil)
	for i := 0; i < len(w.owner.terms); i++ {
		t := w.owner.terms[i]
		state := w.states[i].State(context.Ord)
		if state == nil { /* term doesnt exist in this segment */
			assert2(w.termNotInReader(reader, t), "no termstate found but term exists in reader")
			return nil, nil
		}
		if err = te.SeekExactFromLast(t.Bytes, state); err != nil {
			return nil, err
		}

		var postingsEnum model.DocsAndPositionsEnum
		postingsEnum, err = te.DocsAndPositionsByFlags(liveDocs, nil, model.DOCS_ENUM_FLAG_NONE)
		if err != nil {
			return nil, err
		}

		// PhraseQuery on a field that did not index
		// positions.
		if postingsEnum == nil {
			if err = te.SeekExactFromLast(t.Bytes, state); err != nil {
				panic("termstate found but no term exists in reader")
			}
			// term does exist, but has no positions
			return nil, fmt.Errorf("field \"%s\" was indexed without position data; cannot run PhraseQuery (term=%s)", t.Field, string(t.Bytes))
		}
		var df int
		df, err = te.DocFreq()
		if err != nil {
			return nil, err
		}
		postingsFreqs[i] = newPostingsAndFreq(postingsEnum, df, w.owner.positions[i], t)
	}

	// sort by increasing docFreq order
	if w.owner.slop == 0 {
		util.TimSort(PostingsAndFreqSorter(postingsFreqs))
	}

	var ss SimScorer
	ss, err = w.similarity.SimScorer(w.stats, context)
	if err != nil {
		return nil, err
	}
	if w.owner.slop == 0 { // optimize exact case
		return newExactPhraseScorer(w, postingsFreqs, ss)
	}

	panic("niy")
	return nil, nil
	// return newSloppyPhraseScorer(w, postingsFreqs, w.owner.slop, ss)
}

type PostingsAndFreq struct {
	postings model.DocsAndPositionsEnum
	docFreq  int
	position int32
	terms    []*index.Term
	nTerms   int // for faster comparisons
}

func newPostingsAndFreq(postings model.DocsAndPositionsEnum, docFreq int, position int32, terms ...*index.Term) *PostingsAndFreq {
	var nTerms int
	if terms != nil {
		nTerms = len(terms)
	}
	ans := &PostingsAndFreq{
		postings: postings,
		docFreq:  docFreq,
		position: position,
		nTerms:   nTerms,
	}

	if nTerms > 0 {
		if len(terms) == 1 {
			ans.terms = terms
		} else {
			terms2 := make([]*index.Term, len(terms))
			copy(terms2, terms)
			sort.Sort(index.TermSorter(terms2))
			ans.terms = terms2
		}
	} else {
		ans.terms = nil
	}

	return ans
}

type PostingsAndFreqSorter []*PostingsAndFreq

func (s PostingsAndFreqSorter) Len() int      { return len(s) }
func (s PostingsAndFreqSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PostingsAndFreqSorter) Less(i, j int) bool {
	if s[i].docFreq != s[j].docFreq {
		return (s[i].docFreq - s[j].docFreq) <= 0
	}
	if s[i].position != s[j].position {
		return (s[i].position - s[j].position) <= 0
	}
	if s[i].nTerms != s[j].nTerms {
		return (s[i].nTerms - s[j].nTerms) <= 0
	}
	if s[i].nTerms == 0 {
		return true
	}
	for i := 0; i < len(s[i].terms); i++ {
		return (s[i].terms[i] != (s[j].terms[i]))
	}
	return true
}

func (tw *PhraseWeight) termNotInReader(reader index.IndexReader, term *index.Term) bool {
	n, err := reader.DocFreq(term)
	assert(err == nil)
	return n == 0
}

func (q *PhraseQuery) CreateWeight(searcher *IndexSearcher) (w Weight, err error) {
	return newPhraseWeight(q, searcher)
}

func (q *PhraseQuery) Rewrite(reader index.IndexReader) Query {
	if len(q.terms) == 0 {
		bq := NewBooleanQuery()
		// bq.setBoost(getBoost());
		return bq
	} else if len(q.terms) == 1 {
		tq := NewTermQuery(q.terms[0])
		// tq.setBoost(getBoost());
		return tq
	}
	return q.AbstractQuery.Rewrite(reader)
}

func (q *PhraseQuery) ToString(f string) string {
	var buf bytes.Buffer
	if q.field != "" && q.field != f {
		buf.WriteString(q.field)
		buf.WriteRune(':')
	}

	buf.WriteString("\"")
	pieces := make([]string, q.maxPosition+1)
	for i := 0; i < len(q.terms); i++ {
		pos := q.positions[i]
		s := pieces[pos]
		if s == "" {
			s = string(q.terms[i].Bytes)
		} else {
			s = s + "|" + string(q.terms[i].Bytes)
		}
		pieces[pos] = s
	}
	for i := 0; i < len(pieces); i++ {
		if i > 0 {
			buf.WriteString(" ")
		}
		s := pieces[i]
		if s == "" {
			buf.WriteRune('?')
		} else {
			buf.WriteString(s)
		}
	}

	buf.WriteString("\"")

	if q.slop != 0 {
		buf.WriteRune('~')
		buf.WriteString(strconv.Itoa(q.slop))
	}

	// buf.WriteString(ToStringUtils.boost(getBoost()))

	return buf.String()
}
