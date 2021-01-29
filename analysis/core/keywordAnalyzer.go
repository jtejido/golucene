package core

import (
	. "github.com/jtejido/golucene/core/analysis"
	"io"
)

/**
 * "Tokenizes" the entire stream as a single token. This is useful
 * for data like zip codes, ids, and some product names.
 */
type KeywordAnalyzer struct {
	*AnalyzerImpl
}

func NewKeywordAnalyzer() *KeywordAnalyzer {
	ans := new(KeywordAnalyzer)
	ans.AnalyzerImpl = NewAnalyzer()
	return ans
}

func (a *KeywordAnalyzer) CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents {
	return NewTokenStreamComponentsUnfiltered(NewDefaultKeywordTokenizer(reader))
}
