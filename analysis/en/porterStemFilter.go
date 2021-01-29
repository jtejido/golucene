package en

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
)

type PorterStemFilter struct {
	*TokenFilter
	stemmer     *PorterStemmer
	termAtt     CharTermAttribute
	keywordAttr KeywordAttribute
	input       TokenStream
}

func newPorterStemFilter(in TokenStream) *PorterStemFilter {
	ans := &PorterStemFilter{
		TokenFilter: NewTokenFilter(in),
		stemmer:     newPorterStemmer(),
		input:       in,
	}

	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.keywordAttr = ans.Attributes().Add("KeywordAttribute").(KeywordAttribute)
	return ans
}

func (f *PorterStemFilter) IncrementToken() (bool, error) {
	if v, err := f.input.IncrementToken(); !v {
		return v, err
	}

	if (!f.keywordAttr.IsKeyword()) && f.stemmer.Stem(f.termAtt.Buffer()[:f.termAtt.Length()]) {
		// f.termAtt.CopyBuffer(f.stemmer.ResultBuffer(), 0, f.stemmer.ResultLength())
		f.termAtt.CopyBuffer(f.stemmer.ResultBuffer()[:f.stemmer.ResultLength()])
	}

	return true, nil
}
