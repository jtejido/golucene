package search

import (
	"bytes"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/util"
)

const maxClauseCount = 1024

type BooleanQuery struct {
	*AbstractQuery
	clauses          []*BooleanClause
	disableCoord     bool
	minNrShouldMatch int
}

func NewBooleanQuery() *BooleanQuery {
	return NewBooleanQueryDisableCoord(false)
}

func NewBooleanQueryDisableCoord(disableCoord bool) *BooleanQuery {
	ans := &BooleanQuery{
		disableCoord: disableCoord,
	}
	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}

func (q *BooleanQuery) Add(query Query, occur Occur) {
	q.AddClause(NewBooleanClause(query, occur))
}

func (q *BooleanQuery) AddClause(clause *BooleanClause) {
	assert(len(q.clauses) < maxClauseCount)
	q.clauses = append(q.clauses, clause)
}

type BooleanWeight struct {
	*WeightImpl
	owner        *BooleanQuery
	similarity   Similarity
	weights      []Weight
	maxCoord     int // num optional +num required
	disableCoord bool
}

func newBooleanWeight(owner *BooleanQuery,
	searcher *IndexSearcher, disableCoord bool) (w *BooleanWeight, err error) {

	w = &BooleanWeight{
		owner:        owner,
		similarity:   searcher.similarity,
		disableCoord: disableCoord,
	}
	var subWeight Weight
	for _, c := range owner.clauses {
		if subWeight, err = c.query.CreateWeight(searcher); err != nil {
			return nil, err
		}
		w.weights = append(w.weights, subWeight)
		if !c.IsProhibited() {
			w.maxCoord++
		}
	}
	w.WeightImpl = newWeightImpl(w)
	return w, nil
}

func (w *BooleanWeight) ValueForNormalization() (sum float32) {
	for i, subWeight := range w.weights {
		// call sumOfSquaredWeights for all clauses in case of side effects
		s := subWeight.ValueForNormalization() // sum sub weights
		if !w.owner.clauses[i].IsProhibited() {
			// only add to sum for non-prohibited clauses
			sum += s
		}
	}

	sum *= (w.owner.boost * w.owner.boost) // boost each sub-weight
	return
}

func (w *BooleanWeight) coord(overlap, maxOverlap int) float32 {
	if maxOverlap == 1 {
		return 1
	}
	return w.similarity.Coord(overlap, maxOverlap)
}

func (w *BooleanWeight) Normalize(norm, topLevelBoost float32) {
	topLevelBoost *= w.owner.boost
	for _, subWeight := range w.weights {
		// normalize all clauses, (even if prohibited in case of side effects)
		subWeight.Normalize(norm, topLevelBoost)
	}
}

func (w *BooleanWeight) Explain(context *index.AtomicReaderContext, doc int) (Explanation, error) {
	panic("not implemented yet")
}

func (w *BooleanWeight) BulkScorer(context *index.AtomicReaderContext,
	scoreDocsInOrder bool, acceptDocs util.Bits) (BulkScorer, error) {

	if scoreDocsInOrder || w.owner.minNrShouldMatch > 1 {
		return w.WeightImpl.BulkScorer(context, scoreDocsInOrder, acceptDocs)
	}

	var prohibited, optional []BulkScorer
	for i, subWeight := range w.weights {
		c := w.owner.clauses[i]
		subScorer, err := subWeight.BulkScorer(context, false, acceptDocs)
		if err != nil {
			return nil, err
		}
		if subScorer == nil {
			if c.IsRequired() {
				return nil, nil
			}
		} else if c.IsRequired() {
			return w.WeightImpl.BulkScorer(context, scoreDocsInOrder, acceptDocs)
		} else if c.IsProhibited() {
			prohibited = append(prohibited, subScorer)
		} else {
			optional = append(optional, subScorer)
		}
	}

	return newBooleanScorer(w, w.disableCoord, w.owner.minNrShouldMatch, optional, prohibited, w.maxCoord), nil
}

func (w *BooleanWeight) Scorer(context *index.AtomicReaderContext, acceptDocs util.Bits) (Scorer, error) {
	// initially the user provided value,
	// but if minNrShouldMatch == optional.size(),
	// we will optimize and move these to required, making this 0
	minShouldMatch := w.owner.minNrShouldMatch

	required := make([]Scorer, 0)
	prohibited := make([]Scorer, 0)
	optional := make([]Scorer, 0)
	for i, weight := range w.weights {
		c := w.owner.clauses[i]
		subScorer, err := weight.Scorer(context, acceptDocs)
		if err != nil {
			return nil, err
		}

		if subScorer == nil {
			if c.IsRequired() {
				return nil, nil
			}
		} else if c.IsRequired() {
			required = append(required, subScorer)
		} else if c.IsRequired() {
			prohibited = append(prohibited, subScorer)
		} else {
			optional = append(optional, subScorer)
		}
	}

	// scorer simplifications:

	if len(optional) == minShouldMatch {
		// any optional clauses are in fact required
		required = append(required, optional...)
		optional = make([]Scorer, 0)
		minShouldMatch = 0
	}

	if len(required) == 0 && len(optional) == 0 {
		// no required and optional clauses.
		return nil, nil
	} else if len(optional) < minShouldMatch {
		// either >1 req scorer, or there are 0 req scorers and at least 1
		// optional scorer. Therefore if there are not enough optional scorers
		// no documents will be matched by the query
		return nil, nil
	}

	// three cases: conjunction, disjunction, or mix

	// pure conjunction
	if len(optional) == 0 {
		var s Scorer
		var err error
		if s, err = w.req(required, w.disableCoord); err == nil {
			return w.excl(s, prohibited)
		}

		if err != nil {
			return nil, err
		}
	}

	// pure disjunction
	if len(required) == 0 {
		var s Scorer
		var err error
		if s, err = w.opt(optional, minShouldMatch, w.disableCoord); err == nil {
			return w.excl(s, prohibited)
		}

		if err != nil {
			return nil, err
		}
	}

	// conjunction-disjunction mix:
	// we create the required and optional pieces with coord disabled, and then
	// combine the two: if minNrShouldMatch > 0, then its a conjunction: because the
	// optional side must match. otherwise its required + optional, factoring the
	// number of optional terms into the coord calculation
	var err error
	var req, opt, tmp Scorer

	tmp, err = w.req(required, true)
	if err != nil {
		return nil, err
	}
	req, err = w.excl(tmp, prohibited)
	if err != nil {
		return nil, err
	}
	opt, err = w.opt(optional, minShouldMatch, true)
	if err != nil {
		return nil, err
	}

	// TODO: clean this up: its horrible
	if w.disableCoord {
		if minShouldMatch > 0 {
			return newConjunctionScorerWithCoord(w, []Scorer{req, opt}, 1)
		} else {
			return newReqOptSumScorer(req, opt)
		}
	} else if len(optional) == 1 {
		if minShouldMatch > 0 {
			return newConjunctionScorerWithCoord(w, []Scorer{req, opt}, w.coord(len(required)+1, w.maxCoord))
		} else {
			coordReq := w.coord(len(required), w.maxCoord)
			coordBoth := w.coord(len(required)+1, w.maxCoord)
			return newReqSingleOptScorer(req, opt, coordReq, coordBoth)
		}
	} else {
		if minShouldMatch > 0 {
			return newCoordinatingConjunctionScorer(w, w.coords(), req, len(required), opt)
		} else {
			return newReqMultiOptScorer(req, opt, len(required), w.coords())
		}
	}
}

func (w *BooleanWeight) IsScoresDocsOutOfOrder() bool {
	if w.owner.minNrShouldMatch > 1 {
		// BS2 (in-order) will be used by scorer()
		return false
	}
	optionalCount := 0
	for _, c := range w.owner.clauses {
		if c.IsRequired() {
			// BS2 (in-order) will be used by scorer()
			return false
		} else if !c.IsProhibited() {
			optionalCount++
		}
	}

	if optionalCount == w.owner.minNrShouldMatch {
		return false // BS2 (in-order) will be used, as this means conjunction
	}

	// scorer() will return an out-of-order scorer if requested.
	return true
}

func (w *BooleanWeight) req(required []Scorer, disableCoord bool) (Scorer, error) {
	if len(required) == 1 {
		req := required[0]
		if !w.disableCoord && w.maxCoord > 1 {
			return newBoostedScorer(req, w.coord(1, w.maxCoord))
		} else {
			return req, nil
		}
	} else {
		v := float32(1)
		if !w.disableCoord {
			v = w.coord(len(required), w.maxCoord)
		}
		return newConjunctionScorerWithCoord(w, required, v)
	}
}

func (w *BooleanWeight) excl(main Scorer, prohibited []Scorer) (Scorer, error) {
	if len(prohibited) == 0 {
		return main, nil
	} else if len(prohibited) == 1 {
		return newReqExclScorer(main, prohibited[0])
	} else {
		coords := make([]float32, len(prohibited)+1)
		for i := 0; i < len(coords); i++ {
			coords[i] = 1.
		}
		dss, err := newDisjunctionSumScorer(w, prohibited, coords)
		if err != nil {
			return nil, err
		}
		return newReqExclScorer(main, dss)
	}
}

func (w *BooleanWeight) opt(optional []Scorer, minShouldMatch int, disableCoord bool) (Scorer, error) {
	if len(optional) == 1 {
		opt := optional[0]
		println(opt.Weight())
		if !w.disableCoord && w.maxCoord > 1 {
			return newBoostedScorer(opt, w.coord(1, w.maxCoord))
		} else {
			return opt, nil
		}
	} else {
		var coords []float32
		if w.disableCoord {
			coords = make([]float32, len(optional)+1)
			for i := 0; i < len(coords); i++ {
				coords[i] = 1.
			}
		} else {
			coords = w.coords()
		}
		if minShouldMatch > 1 {
			return newMinShouldMatchSumScorer(w, optional, minShouldMatch, coords)
		} else {
			return newDisjunctionSumScorer(w, optional, coords)
		}
	}
}

func (w *BooleanWeight) coords() []float32 {
	coords := make([]float32, w.maxCoord+1)
	coords[0] = 0
	for i := 1; i < len(coords); i++ {
		coords[i] = w.coord(i, w.maxCoord)
	}
	return coords
}

func (q *BooleanQuery) CreateWeight(searcher *IndexSearcher) (Weight, error) {
	return newBooleanWeight(q, searcher, q.disableCoord)
}

func (q *BooleanQuery) Rewrite(reader index.IndexReader) Query {
	if q.minNrShouldMatch == 0 && len(q.clauses) == 1 {
		panic("not implemented yet")
	}

	var clone *BooleanQuery // recursively rewrite
	for _, c := range q.clauses {
		if query := c.query.Rewrite(reader); query != c.query {
			// clause rewrote: must clone
			if clone == nil {
				// The BooleanQuery clone is lazily initialized so only
				// initialize it if a rewritten clause differs from the
				// original clause (and hasn't been initialized already). If
				// nothing difers, the clone isn't needlessly created
				panic("not implemented yet")
			}
			panic("not implemented yet")
		}
	}
	if clone != nil {
		return clone // some clauses rewrote
	}
	return q
}

func (q *BooleanQuery) ToString(field string) string {
	var buf bytes.Buffer
	needParens := q.Boost() != 1 || q.minNrShouldMatch > 0
	if needParens {
		buf.WriteRune('(')
	}

	for i, c := range q.clauses {
		if c.IsProhibited() {
			buf.WriteRune('-')
		} else if c.IsRequired() {
			buf.WriteRune('+')
		}

		if subQuery := c.query; subQuery != nil {
			if _, ok := subQuery.(*BooleanQuery); ok { // wrap sub-bools in parens
				buf.WriteRune('(')
				buf.WriteString(subQuery.ToString(field))
				buf.WriteRune(')')
			} else {
				buf.WriteString(subQuery.ToString(field))
			}
		} else {
			buf.WriteString("nil")
		}

		if i != len(q.clauses)-1 {
			buf.WriteRune(' ')
		}
	}

	if needParens {
		buf.WriteRune(')')
	}

	if q.minNrShouldMatch > 0 {
		panic("not implemented yet")
	}

	if q.Boost() != 1 {
		panic("not implemented yet")
	}

	return buf.String()
}
