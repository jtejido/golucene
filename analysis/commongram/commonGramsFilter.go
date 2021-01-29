package commongram

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
)

const (
	GRAM_TYPE = "gram"
	SEPARATOR = '_'
)

type CommonGramsFilter struct {
	*TokenFilter
	input           TokenStream
	commonWords     map[string]bool
	buffer          []rune
	termAttribute   CharTermAttribute
	offsetAttribute OffsetAttribute
	posIncAttribute PositionIncrementAttribute
	typeAttribute   TypeAttribute
	posLenAttribute PositionLengthAttribute
	lastStartOffset int
	lastWasCommon   bool
	savedState      *util.AttributeState
}

func NewCommonGramsFilter(in TokenStream, commonWords map[string]bool) *CommonGramsFilter {
	ans := &CommonGramsFilter{
		TokenFilter: NewTokenFilter(in),
		input:       in,
		commonWords: commonWords,
		buffer:      make([]rune, 0),
	}
	ans.termAttribute = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.offsetAttribute = ans.Attributes().Add("OffsetAttribute").(OffsetAttribute)
	ans.posIncAttribute = ans.Attributes().Add("PositionIncrementAttribute").(PositionIncrementAttribute)
	ans.typeAttribute = ans.Attributes().Add("TypeAttribute").(TypeAttribute)
	ans.posLenAttribute = ans.Attributes().Add("PositionLengthAttribute").(PositionLengthAttribute)
	return ans
}

func (f *CommonGramsFilter) IncrementToken() (bool, error) {

	if f.savedState != nil {
		f.Attributes().RestoreState(f.savedState)
		f.savedState = nil
		f.saveTermBuffer()
		return true, nil
	} else {
		ok, err := f.input.IncrementToken()
		if err != nil {
			return false, err
		}

		if !ok {
			return false, nil
		}
	}

	/* We build n-grams before and after stopwords.
	 * When valid, the buffer always contains at least the separator.
	 * If its empty, there is nothing before this stopword.
	 */
	if f.lastWasCommon || (f.isCommon() && len(f.buffer) > 0) {
		f.savedState = f.Attributes().CaptureState()
		f.gramToken()
		return true, nil
	}

	f.saveTermBuffer()
	return true, nil
}

func (f *CommonGramsFilter) Reset() error {
	if err := f.TokenFilter.Reset(); err != nil {
		return err
	}
	f.lastWasCommon = false
	f.savedState = nil
	f.buffer = make([]rune, 0)
	return nil
}

func (f *CommonGramsFilter) isCommon() bool {
	if f.commonWords != nil {
		term := string(f.termAttribute.Buffer()[:f.termAttribute.Length()])
		_, ok := f.commonWords[term]
		return ok
	}
	return false
}

func (f *CommonGramsFilter) saveTermBuffer() {
	f.buffer = make([]rune, 0)
	f.buffer = append(f.buffer, f.termAttribute.Buffer()[:f.termAttribute.Length()]...)
	f.buffer = append(f.buffer, SEPARATOR)
	f.lastStartOffset = f.offsetAttribute.StartOffset()
	f.lastWasCommon = f.isCommon()
}

func (f *CommonGramsFilter) gramToken() {
	f.buffer = append(f.buffer, f.termAttribute.Buffer()[:f.termAttribute.Length()]...)
	endOffset := f.offsetAttribute.EndOffset()

	f.Attributes().Clear()

	length := len(f.buffer)
	termText := f.termAttribute.Buffer()
	if length > len(termText) {
		termText = f.termAttribute.ResizeBuffer(length)
	}
	copy(termText[0:], f.buffer[:length-1])
	f.termAttribute.SetLength(length)
	f.posIncAttribute.SetPositionIncrement(0)
	f.posLenAttribute.SetPositionLength(2) // bigram
	f.offsetAttribute.SetOffset(f.lastStartOffset, endOffset)
	f.typeAttribute.SetType(GRAM_TYPE)
	f.buffer = make([]rune, 0)
}
