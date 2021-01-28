package standard

import (
	. "github.com/jtejido/golucene/core/analysis"
	"github.com/jtejido/golucene/core/util"
)

// standard/StandardFilter.java

/* Normalizes tokens extracted with StandardTokenizer */
type StandardFilter struct {
	*TokenFilter
	matchVersion util.Version
	input        TokenStream
}

func NewStandardFilter(matchVersion util.Version, in TokenStream) *StandardFilter {
	return &StandardFilter{
		TokenFilter:  NewTokenFilter(in),
		matchVersion: matchVersion,
		input:        in,
	}
}

func (f *StandardFilter) IncrementToken() (bool, error) {
	assert(f.matchVersion.OnOrAfter(util.VERSION_31))
	return f.input.IncrementToken()
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
