package classic

import (
	"errors"
	// "fmt"
	"github.com/jtejido/golucene/core/analysis"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/util"
	"strings"
)

type Operator int

var (
	OP_OR  = Operator(1)
	OP_AND = Operator(2)
)

type QueryParser struct {
	*QueryParserBase

	token_source           *TokenManager
	token                  *Token // current token
	jj_nt                  *Token // next token
	jj_ntk                 int
	jj_scanpos, jj_lastpos *Token
	jj_la                  int
	jj_gen                 int
	jj_la1                 []int
	jj_kind                int

	jj_2_rtns []*JJCalls
	jj_rescan bool
	jj_gc     int
}

func NewQueryParser(matchVersion util.Version, f string, a analysis.Analyzer) *QueryParser {
	qp := &QueryParser{
		token_source: newTokenManager(newFastCharStream(strings.NewReader(""))),
		jj_la1:       make([]int, 21),
		jj_2_rtns:    make([]*JJCalls, 1),
		jj_kind:      -1,
	}
	qp.QueryParserBase = newQueryParserBase(qp)
	qp.ReInit(newFastCharStream(strings.NewReader("")))
	// base
	qp.analyzer = a
	qp.field = f
	qp.autoGeneratePhraseQueries = !matchVersion.OnOrAfter(util.VERSION_31)
	return qp
}

func (qp *QueryParser) conjunction() (int, error) {
	ret := CONJ_NONE
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case AND, OR:
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case AND:
			qp.jj_consume_token(AND)
			ret = CONJ_AND
			break
		case OR:
			qp.jj_consume_token(OR)
			ret = CONJ_OR
			break
		default:
			qp.jj_la1[0] = qp.jj_gen
			qp.jj_consume_token(-1)
			return 0, errors.New("parse error")
		}
	default:
		qp.jj_la1[1] = qp.jj_gen
	}
	return ret, nil
}

func (qp *QueryParser) modifiers() (ret int, err error) {
	ret = MOD_NONE
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case NOT, PLUS, MINUS:
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case PLUS:
			qp.jj_consume_token(PLUS)
			ret = MOD_REQ
			break
		case MINUS:
			qp.jj_consume_token(MINUS)
			ret = MOD_NOT
			break
		case NOT:
			qp.jj_consume_token(NOT)
			ret = MOD_NOT
			break
		default:
			qp.jj_la1[2] = qp.jj_gen
			qp.jj_consume_token(-1)
			return 0, errors.New("parse error")
		}
		break
	default:
		qp.jj_la1[3] = qp.jj_gen
	}
	return
}

func (qp *QueryParser) TopLevelQuery(field string) (q search.Query, err error) {
	if q, err = qp.Query(field); err != nil {
		return nil, err
	}
	_, err = qp.jj_consume_token(0)
	return q, err
}

func (qp *QueryParser) Query(field string) (q search.Query, err error) {
	var clauses []*search.BooleanClause
	var conj, mods int
	if mods, err = qp.modifiers(); err != nil {
		return nil, err
	}
	if q, err = qp.clause(field); err != nil {
		return nil, err
	}
	clauses = qp.addClause(clauses, CONJ_NONE, mods, q)
	var firstQuery search.Query
	if mods == MOD_NONE {
		firstQuery = q
	}
label_1:
	for {
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case AND, OR, NOT, PLUS, MINUS, BAREOPER, LPAREN, STAR, QUOTED,
			TERM, PREFIXTERM, WILDTERM, REGEXPTERM, RANGEIN_START,
			RANGEEX_START, NUMBER:
			break
		default:
			qp.jj_la1[4] = qp.jj_gen
			break label_1
		}

		if conj, err = qp.conjunction(); err != nil {
			return nil, err
		}
		if mods, err = qp.modifiers(); err != nil {
			return nil, err
		}
		if q, err = qp.clause(field); err != nil {
			return nil, err
		}
		clauses = qp.addClause(clauses, conj, mods, q)
	}
	if len(clauses) == 1 && firstQuery != nil {
		return firstQuery, nil
	} else {
		return qp.booleanQuery(clauses)
	}
}

func (qp *QueryParser) clause(field string) (q search.Query, err error) {
	var fieldToken *Token = nil
	if qp.jj_2_1(2) {
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case TERM:
			fieldToken, err = qp.jj_consume_token(TERM)
			if err != nil {
				return nil, err
			}
			qp.jj_consume_token(COLON)
			field, err = qp.discardEscapeChar(fieldToken.image)
			if err != nil {
				return nil, err
			}
			break
		case STAR:
			qp.jj_consume_token(STAR)
			qp.jj_consume_token(COLON)
			field = "*"
			break
		default:
			qp.jj_la1[5] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}
	}
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	var boost *Token
	switch qp.jj_ntk {
	case BAREOPER, STAR, QUOTED, TERM, PREFIXTERM, WILDTERM,
		REGEXPTERM, RANGEIN_START, RANGEEX_START, NUMBER:
		if q, err = qp.term(field); err != nil {
			return nil, err
		}
		break
	case LPAREN:
		qp.jj_consume_token(LPAREN)
		q, err = qp.Query(field)
		if err != nil {
			return nil, err
		}
		qp.jj_consume_token(RPAREN)
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case CARAT:
			qp.jj_consume_token(CARAT)
			boost, err = qp.jj_consume_token(NUMBER)
			if err != nil {
				return nil, err
			}
			break
		default:
			qp.jj_la1[6] = qp.jj_gen

		}
		break
	default:
		qp.jj_la1[7] = qp.jj_gen
		if _, err = qp.jj_consume_token(-1); err != nil {
			return nil, err
		}
		return nil, errors.New("parse error")
	}
	return qp.handleBoost(q, boost), nil
}

func (qp *QueryParser) term(field string) (q search.Query, err error) {
	var term, boost, fuzzySlop, goop1, goop2 *Token
	var prefix, wildcard, fuzzy, regexp, startInc, endInc bool
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case BAREOPER, STAR, TERM, PREFIXTERM, WILDTERM, REGEXPTERM, NUMBER:
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case TERM:
			if term, err = qp.jj_consume_token(TERM); err != nil {
				return nil, err
			}
			break
		case STAR:
			if term, err = qp.jj_consume_token(STAR); err != nil {
				return nil, err
			}
			wildcard = true
			break
		case PREFIXTERM:
			if term, err = qp.jj_consume_token(PREFIXTERM); err != nil {
				return nil, err
			}
			prefix = true
			break
		case WILDTERM:
			if term, err = qp.jj_consume_token(WILDTERM); err != nil {
				return nil, err
			}
			wildcard = true
			break
		case REGEXPTERM:
			if term, err = qp.jj_consume_token(REGEXPTERM); err != nil {
				return nil, err
			}
			regexp = true
			break
		case NUMBER:
			if term, err = qp.jj_consume_token(NUMBER); err != nil {
				return nil, err
			}
			break
		case BAREOPER:
			if term, err = qp.jj_consume_token(BAREOPER); err != nil {
				return nil, err
			}
			term.image = term.image[:1]
			break
		default:
			qp.jj_la1[8] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case FUZZY_SLOP:
			if fuzzySlop, err = qp.jj_consume_token(FUZZY_SLOP); err != nil {
				return nil, err
			}
			fuzzy = true
			break
		default:
			qp.jj_la1[9] = qp.jj_gen
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case CARAT:
			qp.jj_consume_token(CARAT)
			if boost, err = qp.jj_consume_token(NUMBER); err != nil {
				return nil, err
			}
			if qp.jj_ntk == -1 {
				qp.get_jj_ntk()
			}
			switch qp.jj_ntk {
			case FUZZY_SLOP:
				if fuzzySlop, err = qp.jj_consume_token(FUZZY_SLOP); err != nil {
					return nil, err
				}
				fuzzy = true
				break
			default:
				qp.jj_la1[10] = qp.jj_gen
			}
			break
		default:
			qp.jj_la1[11] = qp.jj_gen
		}
		if q, err = qp.handleBareTokenQuery(field, term, fuzzySlop, prefix, wildcard, fuzzy, regexp); err != nil {
			return nil, err
		}
		break
	case RANGEIN_START, RANGEEX_START:
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case RANGEIN_START:
			qp.jj_consume_token(RANGEIN_START)
			startInc = true
			break
		case RANGEEX_START:
			qp.jj_consume_token(RANGEEX_START)
			break
		default:
			qp.jj_la1[12] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}

		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case RANGE_GOOP:
			if goop1, err = qp.jj_consume_token(RANGE_GOOP); err != nil {
				return nil, err
			}
			break
		case RANGE_QUOTED:
			if goop1, err = qp.jj_consume_token(RANGE_QUOTED); err != nil {
				return nil, err
			}
			break
		default:
			qp.jj_la1[13] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case RANGE_TO:
			qp.jj_consume_token(RANGE_TO)
			break
		default:
			qp.jj_la1[14] = qp.jj_gen

		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case RANGE_GOOP:
			if goop2, err = qp.jj_consume_token(RANGE_GOOP); err != nil {
				return nil, err
			}
		case RANGE_QUOTED:
			if goop2, err = qp.jj_consume_token(RANGE_QUOTED); err != nil {
				return nil, err
			}
		default:
			qp.jj_la1[15] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}

		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case RANGEIN_END:
			qp.jj_consume_token(RANGEIN_END)
			endInc = true
			break
		case RANGEEX_END:
			qp.jj_consume_token(RANGEEX_END)
			break
		default:
			qp.jj_la1[16] = qp.jj_gen
			qp.jj_consume_token(-1)
			return nil, errors.New("parse error")
		}

		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case CARAT:
			qp.jj_consume_token(CARAT)
			if boost, err = qp.jj_consume_token(NUMBER); err != nil {
				return nil, err
			}
			break
		default:
			qp.jj_la1[17] = qp.jj_gen

		}
		var startOpen, endOpen bool
		if goop1.kind == RANGE_QUOTED {
			goop1.image = goop1.image[1 : len(goop1.image)-1]
		} else if "*" == goop1.image {
			startOpen = true
		}
		if goop2.kind == RANGE_QUOTED {
			goop2.image = goop2.image[1 : len(goop2.image)-1]
		} else if "*" == goop2.image {
			endOpen = true
		}
		var s1, s2 string
		if !startOpen {
			if s1, err = qp.discardEscapeChar(goop1.image); err != nil {
				return nil, err
			}
		}

		if !endOpen {
			if s2, err = qp.discardEscapeChar(goop2.image); err != nil {
				return nil, err
			}
		}
		q = qp.rangeQuery(field, s1, s2, startInc, endInc)
		break
	case QUOTED:
		if term, err = qp.jj_consume_token(QUOTED); err != nil {
			return nil, err
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case FUZZY_SLOP:
			if fuzzySlop, err = qp.jj_consume_token(FUZZY_SLOP); err != nil {
				return nil, err
			}
			break
		default:
			qp.jj_la1[18] = qp.jj_gen
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case CARAT:
			qp.jj_consume_token(CARAT)
			if boost, err = qp.jj_consume_token(NUMBER); err != nil {
				return nil, err
			}
			break
		default:
			qp.jj_la1[19] = qp.jj_gen
		}
		q, err = qp.handleQuotedTerm(field, term, fuzzySlop)
		if err != nil {
			return nil, err
		}
		break
	default:
		qp.jj_la1[20] = qp.jj_gen
		qp.jj_consume_token(-1)
		return nil, errors.New("parse error")
	}
	return qp.handleBoost(q, boost), nil
}

// L473
func (qp *QueryParser) jj_2_1(xla int) (ok bool) {
	qp.jj_la = xla
	qp.jj_lastpos = qp.token
	qp.jj_scanpos = qp.token
	defer func() {
		// Disable following recover() to deal with dev panic
		if err := recover(); err == lookAheadSuccess {
			ok = true
		}
		qp.jj_save(0, xla)
	}()
	return !qp.jj_3_1()
}

func (qp *QueryParser) jj_3R_2() bool {
	return qp.jj_scan_token(TERM) ||
		qp.jj_scan_token(COLON)
}

func (qp *QueryParser) jj_3_1() bool {
	xsp := qp.jj_scanpos
	if qp.jj_3R_2() {
		qp.jj_scanpos = xsp
		if qp.jj_3R_3() {
			return true
		}
	}
	return false
}

func (qp *QueryParser) jj_3R_3() bool {
	return qp.jj_scan_token(STAR) ||
		qp.jj_scan_token(COLON)
}

// L540
func (qp *QueryParser) ReInit(stream CharStream) {
	qp.token_source.ReInit(stream)
	qp.token = new(Token)
	qp.jj_ntk = -1
	qp.jj_gen = 0
	for i, _ := range qp.jj_la1 {
		qp.jj_la1[i] = -1
	}
	for i, _ := range qp.jj_2_rtns {
		qp.jj_2_rtns[i] = new(JJCalls)
	}
}

// L569
func (qp *QueryParser) jj_consume_token(kind int) (*Token, error) {
	oldToken := qp.token
	if qp.token.next != nil {
		qp.token = qp.token.next
	} else {
		qp.token.next = qp.token_source.nextToken()
		qp.token = qp.token.next
	}
	qp.jj_ntk = -1
	if qp.token.kind == kind {
		qp.jj_gen++
		if qp.jj_gc++; qp.jj_gc > 100 {
			qp.jj_gc = 0
			for i := 0; i < len(qp.jj_2_rtns); i++ {
				c := qp.jj_2_rtns[i]
				for c != nil {
					if c.gen < qp.jj_gen {
						c.first = nil
					}
					c = c.next
				}
			}
		}
		return qp.token, nil
	}
	qp.token = oldToken
	qp.jj_kind = kind
	panic("not implemented yet")
}

type LookAheadSuccess bool

var lookAheadSuccess = LookAheadSuccess(true)

func (qp *QueryParser) jj_scan_token(kind int) bool {
	if qp.jj_scanpos == qp.jj_lastpos {
		qp.jj_la--
		if qp.jj_scanpos.next == nil {
			nextToken := qp.token_source.nextToken()
			qp.jj_scanpos.next = nextToken
			qp.jj_scanpos = nextToken
			qp.jj_lastpos = nextToken
		} else {
			qp.jj_scanpos = qp.jj_scanpos.next
			qp.jj_lastpos = qp.jj_scanpos.next
		}
	} else {
		qp.jj_scanpos = qp.jj_scanpos.next
	}
	if qp.jj_rescan {
		panic("niy")
	}
	if qp.jj_scanpos.kind != kind {
		return true
	}
	if qp.jj_la == 0 && qp.jj_scanpos == qp.jj_lastpos {
		panic(lookAheadSuccess)
	}
	return false
}

// L636
func (qp *QueryParser) get_jj_ntk() int {
	if qp.jj_nt = qp.token.next; qp.jj_nt == nil {
		qp.token.next = qp.token_source.nextToken()
		qp.jj_ntk = qp.token.next.kind
	} else {
		qp.jj_ntk = qp.jj_nt.kind
	}
	return qp.jj_ntk
}

// L738

func (qp *QueryParser) jj_save(index, xla int) {
	p := qp.jj_2_rtns[index]
	for p.gen > qp.jj_gen {
		if p.next == nil {
			p = new(JJCalls)
			p.next = p
			break
		}
		p = p.next
	}
	p.gen = qp.jj_gen + xla - qp.jj_la
	p.first = qp.token
	p.arg = xla
}

type JJCalls struct {
	gen   int
	first *Token
	arg   int
	next  *JJCalls
}
