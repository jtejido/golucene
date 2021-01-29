package core

import (
	. "github.com/jtejido/golucene/analysis/util"
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
)

/**
 * Normalizes token text to UPPER CASE.
 * <a name="version"/>
 * <p>You may specify the {@link Version}
 * compatibility when creating UpperCaseFilter
 *
 * <p><b>NOTE:</b> In Unicode, this transformation may lose information when the
 * upper case character represents more than one lower case character. Use this filter
 * when you require uppercase tokens.  Use the {@link LowerCaseFilter} for
 * general search matching
 */
type UpperCaseFilter struct {
	*TokenFilter
	input     TokenStream
	charUtils *CharacterUtils
	termAtt   CharTermAttribute
}

/* Create a new UpperCaseFilter, that normalizes token text to upper case. */
func NewUpperCaseFilter(matchVersion util.Version, in TokenStream) *UpperCaseFilter {
	ans := &UpperCaseFilter{
		TokenFilter: NewTokenFilter(in),
		input:       in,
		charUtils:   GetCharacterUtils(matchVersion),
	}
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	return ans
}

func (f *UpperCaseFilter) IncrementToken() (bool, error) {
	ok, err := f.input.IncrementToken()
	if err != nil {
		return false, err
	}
	if ok {
		f.charUtils.ToUpperCase(f.termAtt.Buffer()[:f.termAtt.Length()])
		return true, nil
	}
	return false, nil
}
