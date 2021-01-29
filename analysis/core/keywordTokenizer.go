package core

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"io"
)

const (
	DEFAULT_BUFFER_SIZE = 256
)

/**
 * Emits the entire input as a single token.
 */
type KeywordTokenizer struct {
	*Tokenizer
	done        bool
	finalOffset int
	termAtt     CharTermAttribute
	offsetAtt   OffsetAttribute
}

func NewDefaultKeywordTokenizer(input io.RuneReader) *KeywordTokenizer {
	return NewKeywordTokenizer(input, DEFAULT_BUFFER_SIZE)
}

func NewKeywordTokenizer(input io.RuneReader, bufferSize int) *KeywordTokenizer {
	if bufferSize <= 0 {
		panic("bufferSize must be > 0")
	}

	ans := &KeywordTokenizer{
		Tokenizer: NewTokenizer(input),
	}
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.termAtt.ResizeBuffer(bufferSize)
	ans.offsetAtt = ans.Attributes().Add("OffsetAttribute").(OffsetAttribute)
	return ans
}

func readRunes(r io.RuneReader, buffer []rune) (int, error) {
	for i, _ := range buffer {
		ch, _, err := r.ReadRune()
		if err != nil {
			return i, err
		}
		buffer[i] = ch
	}
	return len(buffer), nil
}

func (t *KeywordTokenizer) IncrementToken() (bool, error) {
	if !t.done {
		t.Attributes().Clear()
		t.done = true
		var upto int
		buffer := t.termAtt.Buffer()
		for {
			var length int
			var err error
			if length, err = readRunes(t.Input, buffer[upto:]); err != nil && err != io.EOF {
				return false, err
			}

			if length == -1 {
				break
			}
			upto += length
			if upto == len(buffer) {
				buffer = t.termAtt.ResizeBuffer(1 + len(buffer))
			}
		}
		t.termAtt.SetLength(upto)
		t.finalOffset = t.CorrectOffset(upto)
		t.offsetAtt.SetOffset(t.CorrectOffset(0), t.finalOffset)
		return true, nil
	}
	return false, nil
}

func (t *KeywordTokenizer) End() error {
	err := t.Tokenizer.End()
	if err == nil {
		t.offsetAtt.SetOffset(t.finalOffset, t.finalOffset)
	}
	return nil
}

func (t *KeywordTokenizer) Reset() error {
	if err := t.Tokenizer.Reset(); err != nil {
		return err
	}
	t.done = false
	return nil
}
