package en

import (
	. "github.com/jtejido/golucene/analysis/core"
	. "github.com/jtejido/golucene/analysis/standard"
	. "github.com/jtejido/golucene/analysis/util"
	. "github.com/jtejido/golucene/core/analysis"
	"github.com/jtejido/golucene/core/util"
	"io"
)

/* An unmodifiable set containing some common English words that are not usually useful for searching. */
var ENGLISH_STOP_WORDS_SET = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "but": true, "by": true,
	"for": true, "if": true, "in": true, "into": true, "is": true, "it": true,
	"no": true, "not": true, "of": true, "on": true, "or": true, "such": true,
	"that": true, "the": true, "their": true, "then": true, "there": true, "these": true,
	"they": true, "this": true, "to": true, "was": true, "will": true, "with": true,
}

var STOP_WORDS_SET = ENGLISH_STOP_WORDS_SET

type EnglishAnalyzer struct {
	*StopwordAnalyzerBase
	stopWordSet      map[string]bool
	stemExclusionSet map[string]bool
}

/* Builds an analyzer with the given stop words. */
func NewEnglishAnalyzerWithStopWords(stopWords map[string]bool) *EnglishAnalyzer {
	return NewEnglishAnalyzerWithStopWordsAndStemExclusion(STOP_WORDS_SET, make(map[string]bool))
}

/* Builds an analyzer with the given stop words. */
func NewEnglishAnalyzerWithStopWordsAndStemExclusion(stopWords, stemExclusionSet map[string]bool) *EnglishAnalyzer {
	ans := &EnglishAnalyzer{
		stopWordSet:      stopWords,
		stemExclusionSet: stemExclusionSet,
	}
	ans.StopwordAnalyzerBase = NewStopwordAnalyzerBaseWithStopWords(stopWords)
	ans.Spi = ans
	return ans
}

/* Buils an analyzer with the default stop words (STOP_WORDS_SET). */
func NewEnglishAnalyzer() *EnglishAnalyzer {
	return NewEnglishAnalyzerWithStopWords(STOP_WORDS_SET)
}

func (a *EnglishAnalyzer) CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents {
	version := a.Version()
	src := NewStandardTokenizer(version, reader)
	var result TokenStream = NewStandardFilter(version, src)
	assert(version.OnOrAfter(util.VERSION_36))
	result = newEnglishPossessiveFilter(version, result)
	result = NewLowerCaseFilter(version, result)
	result = NewStopFilter(version, result, a.stopWordSet)
	result = newPorterStemFilter(result)

	return NewTokenStreamComponents(src, result)
}
