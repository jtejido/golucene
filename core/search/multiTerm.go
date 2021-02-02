package search

import (
	"github.com/jtejido/golucene/core/index"
	. "github.com/jtejido/golucene/core/index/model"
	"github.com/jtejido/golucene/core/util"
)

type RewriteMethod interface {
	Rewrite(reader index.IndexReader, query MultiTermQuery) Query
	TermsEnum(query MultiTermQuery, terms Terms, atts *util.AttributeSource) (TermsEnum, error)
}

type RewriteMethodImpl struct {
	spi RewriteMethod
}

func newRewriteMethodImpl(spi RewriteMethod) *RewriteMethodImpl {
	return &RewriteMethodImpl{spi}
}

func (r *RewriteMethodImpl) TermsEnum(query MultiTermQuery, terms Terms, atts *util.AttributeSource) (TermsEnum, error) {
	return query.TermsEnum(terms, atts) // allow RewriteMethod subclasses to pull a TermsEnum from the MTQ
}

type MultiTermQuery interface {
	Query
	/** Construct the enumeration to be used, expanding the
	 *  pattern term.  This method should only be called if
	 *  the field exists (ie, implementations can assume the
	 *  field does exist).  This method should not return null
	 *  (should instead return {@link TermsEnum#EMPTY} if no
	 *  terms match).  The TermsEnum must already be
	 *  positioned to the first matching term.
	 * The given {@link AttributeSource} is passed by the {@link RewriteMethod} to
	 * provide attributes, the rewrite method uses to inform about e.g. maximum competitive boosts.
	 * This is currently only used by {@link TopTermsRewrite}
	 */
	TermsEnum(terms Terms, atts *util.AttributeSource) (TermsEnum, error)

	/**
	 * To rewrite to a simpler form, instead return a simpler
	 * enum from {@link #getTermsEnum(Terms, AttributeSource)}.  For example,
	 * to rewrite to a single term, return a {@link SingleTermsEnum}
	 */
	Rewrite(reader index.IndexReader) Query
	Field() string
	RewriteMethod() RewriteMethod
}

type MultiTermQuerySPI interface {
	TermsEnum(terms Terms, atts *util.AttributeSource) (TermsEnum, error)
}

type AbstractMultiTermQuery struct {
	*AbstractQuery
	spi           MultiTermQuerySPI
	field         string
	rewriteMethod RewriteMethod
}

func newAbstractMultiTermQuery(spi MultiTermQuerySPI, field string) *AbstractMultiTermQuery {
	assert(field != "")
	ans := &AbstractMultiTermQuery{
		spi:   spi,
		field: field,
	}

	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}

func (q *AbstractMultiTermQuery) Rewrite(r index.IndexReader) Query {
	return q.rewriteMethod.Rewrite(r, q)
}

func (q *AbstractMultiTermQuery) TermsEnum(terms Terms, atts *util.AttributeSource) (TermsEnum, error) {
	return q.spi.TermsEnum(terms, atts)
}

func (q *AbstractMultiTermQuery) Field() string {
	return q.field
}

func (q *AbstractMultiTermQuery) RewriteMethod() RewriteMethod {
	return q.rewriteMethod
}
