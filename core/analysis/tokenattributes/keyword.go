package tokenattributes

import (
	"github.com/jtejido/golucene/core/util"
)

// This attribute can be used to mark a token as a keyword. Keyword aware TokenStreams can decide
// to modify a token based on the return value of isKeyword() if the token is modified.
// Stemming filters for instance can use this attribute to conditionally skip a term if isKeyword() returns true.
type KeywordAttribute interface {
	// Returns true if the current token is a keyword, otherwise false.
	IsKeyword() bool

	// Marks the current token as keyword if set to true.
	SetKeyword(isKeyword bool)
}

/* Default implementation of CharTermAttribute. */
type KeywordAttributeImpl struct {
	keyword bool
}

func newKeywordAttributeImpl() *KeywordAttributeImpl {
	return new(KeywordAttributeImpl)
}

func (a *KeywordAttributeImpl) Clear() {
	a.keyword = false
}

func (a *KeywordAttributeImpl) CopyTo(target util.AttributeImpl) {
	attr := target.(KeywordAttribute)
	attr.SetKeyword(a.keyword)
}

func (a *KeywordAttributeImpl) IsKeyword() bool {
	return a.keyword
}

func (a *KeywordAttributeImpl) SetKeyword(isKeyword bool) {
	a.keyword = isKeyword
}

func (a *KeywordAttributeImpl) Clone() util.AttributeImpl {
	return &KeywordAttributeImpl{
		keyword: a.keyword,
	}
}

func (a *KeywordAttributeImpl) Interfaces() []string {
	return []string{"KeywordAttributeImpl"}
}
