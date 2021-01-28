package core

import (
	. "github.com/jtejido/golucene/analysis/util"
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
)

// core/StopFilter.java

/*
Removes stop words from a token stream.

You may specify the Version
compatibility when creating StopFilter:

	- As of 3.1, StopFilter correctly handles Unicode 4.0 supplementary
	characters in stopwords and position increments are preserved
*/
type StopFilter struct {
	*FilteringTokenFilter
	stopWords map[string]bool
	termAtt   CharTermAttribute
}

/*
Constructs a filter which removes words from the input TokenStream
that are named in the Set.
*/
func NewStopFilter(matchVersion util.Version,
	in TokenStream, stopWords map[string]bool) *StopFilter {

	ans := &StopFilter{stopWords: stopWords}
	ans.FilteringTokenFilter = NewFilteringTokenFilter(ans, matchVersion, in)
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	return ans
}

func (f *StopFilter) Accept() bool {
	term := string(f.termAtt.Buffer()[:f.termAtt.Length()])
	_, ok := f.stopWords[term]
	return !ok
}
