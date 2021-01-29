package en

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
)

// en/EnglishPossessiveFilter.java
type EnglishPossessiveFilter struct {
	*TokenFilter
	termAtt      CharTermAttribute
	matchVersion util.Version
	input        TokenStream
}

func newEnglishPossessiveFilter(matchVersion util.Version, in TokenStream) *EnglishPossessiveFilter {
	ans := &EnglishPossessiveFilter{
		TokenFilter:  NewTokenFilter(in),
		matchVersion: matchVersion,
		input:        in,
	}

	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	return ans
}

func (f *EnglishPossessiveFilter) IncrementToken() (bool, error) {
	if v, err := f.input.IncrementToken(); !v {
		return v, err
	}

	buffer := f.termAtt.Buffer()
	bufferLength := f.termAtt.Length()

	assert(f.matchVersion.OnOrAfter(util.VERSION_36))

	if bufferLength >= 2 &&
		(buffer[bufferLength-2] == '\'' || (f.matchVersion.OnOrAfter(util.VERSION_36) && (buffer[bufferLength-2] == '\u2019' || buffer[bufferLength-2] == '\uFF07'))) &&
		(buffer[bufferLength-1] == 's' || buffer[bufferLength-1] == 'S') {
		f.termAtt.SetLength(bufferLength - 2) // Strip last 2 characters off
	}

	return true, nil
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
