package hunspell

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
	"sort"
)

// hunspell/HunspellStemFilter.java
type HunspellStemFilter struct {
	*TokenFilter
	termAtt            CharTermAttribute
	posIncAtt          PositionIncrementAttribute
	keywordAtt         KeywordAttribute
	stemmer            *Stemmer
	buffer             []*util.CharsRef
	savedState         *util.AttributeState
	dedup, longestOnly bool
	input              TokenStream
}

func newHunspellStemFilter(in TokenStream, dictionary *Dictionary, dedup, longestOnly bool) *HunspellStemFilter {
	ans := &HunspellStemFilter{
		TokenFilter: NewTokenFilter(in),
		dedup:       dedup && longestOnly == false,
		longestOnly: longestOnly,
		stemmer:     newStemmer(dictionary),
		input:       in,
	}

	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.posIncAtt = ans.Attributes().Add("PositionIncrementAttribute").(PositionIncrementAttribute)
	ans.keywordAttr = ans.Attributes().Add("KeywordAttribute").(KeywordAttribute)
	return ans
}

func (f *HunspellStemFilter) IncrementToken() (bool, error) {
	if f.buffer != nil && len(f.buffer) > 0 {
		nextStem := f.buffer[0]
		copy(f.buffer, f.buffer[1:])
		f.buffer[len(f.buffer)-1] = nil
		f.buffer = f.buffer[:len(f.buffer)-1]
		f.Attributes().RestoreState(f.savedState)
		f.posIncAtt.SetPositionIncrement(0)
		f.termAtt.Clear()
		f.termAtt.AppendCharsRef(nextStem)
		return true, nil
	}

	if v, err := f.input.IncrementToken(); !v {
		return v, err
	}

	if f.keywordAtt.IsKeyword() {
		return true, nil
	}

	if f.dedup {
		f.buffer = f.stemmer.UniqueStems(f.termAtt.Buffer(), f.termAtt.Length())
	} else {
		f.buffer = f.stemmer.Stem(f.termAtt.Buffer(), f.termAtt.Length())
	}

	if len(f.buffer) == 0 { // we do not know this word, return it unchanged
		return true, nil
	}

	if f.longestOnly && len(f.buffer) > 1 {
		sort.Sort(lengthComparator(f.buffer))
	}

	stem := f.buffer[0]
	copy(f.buffer, f.buffer[1:])
	f.buffer[len(f.buffer)-1] = nil
	f.buffer = f.buffer[:len(f.buffer)-1]
	f.termAtt.Clear()
	f.termAtt.AppendCharsRef(stem)

	if f.longestOnly {
		f.buffer = nil
	} else {
		if len(f.buffer) > 0 {
			savedState = f.Attributes().CaptureState()
		}
	}

	return true, nil
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
