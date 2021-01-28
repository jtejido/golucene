package standard

import (
	. "github.com/jtejido/golucene/analysis/core"
	. "github.com/jtejido/golucene/analysis/util"
	. "github.com/jtejido/golucene/core/analysis"
	"io"
)

// standard/StandardAnalyzer.java

/* Default maximum allowed token length */
const DEFAULT_MAX_TOKEN_LENGTH = 255

/* An unmodifiable set containing some common English words that are not usually useful for searching. */
var STANDARD_STOP_WORDS_SET = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "but": true, "by": true,
	"for": true, "if": true, "in": true, "into": true, "is": true, "it": true,
	"no": true, "not": true, "of": true, "on": true, "or": true, "such": true,
	"that": true, "the": true, "their": true, "then": true, "there": true, "these": true,
	"they": true, "this": true, "to": true, "was": true, "will": true, "with": true,
}

/* An unmodifiable set containing some common English words that are usually not useful for searching */
var STD_STOP_WORDS_SET = STANDARD_STOP_WORDS_SET

/*
Filters StandardTokenizer with StandardFilter, LowerCaseFilter and
StopFilter, using a list of English stop words.

You may specify the Version
compatibility when creating StandardAnalyzer:

	- GoLucene supports 4.5+ only.
*/
type StandardAnalyzer struct {
	*StopwordAnalyzerBase
	stopWordSet    map[string]bool
	maxTokenLength int
}

/* Builds an analyzer with the given stop words. */
func NewStandardAnalyzerWithStopWords(stopWords map[string]bool) *StandardAnalyzer {
	ans := &StandardAnalyzer{
		stopWordSet:    stopWords,
		maxTokenLength: DEFAULT_MAX_TOKEN_LENGTH,
	}
	ans.StopwordAnalyzerBase = NewStopwordAnalyzerBaseWithStopWords(stopWords)
	ans.Spi = ans
	return ans
}

func NewStandardAnalyzer() *StandardAnalyzer {
	return NewStandardAnalyzerWithStopWords(STD_STOP_WORDS_SET)
}

func (a *StandardAnalyzer) CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents {
	version := a.Version()
	src := NewStandardTokenizer(version, reader)
	src.maxTokenLength = a.maxTokenLength
	var tok TokenStream = NewStandardFilter(version, src)
	tok = NewLowerCaseFilter(version, tok)
	tok = NewStopFilter(version, tok, a.stopWordSet)
	ans := NewTokenStreamComponents(src, tok)
	super := ans.SetReader
	ans.SetReader = func(reader io.RuneReader) error {
		src.maxTokenLength = a.maxTokenLength
		return super(reader)
	}
	return ans
}
