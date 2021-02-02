package standard

import (
	. "github.com/jtejido/golucene/core/analysis"
	. "github.com/jtejido/golucene/core/analysis/tokenattributes"
	"github.com/jtejido/golucene/core/util"
	"io"
)

// standard/StandardTokenizer.java

const (
	/** Alpha/numeric token type */
	ALPHANUM = 0
	/** Numeric token type */
	NUM = 1
	/** Southeast Asian token type */
	SOUTHEAST_ASIAN = 2
	/** Ideographic token type */
	IDEOGRAPHIC = 3
	/** Hiragana token type */
	HIRAGANA = 4
	/** Katakana token type */
	KATAKANA = 5
	/** Hangul token type */
	HANGUL = 6
	/** Emoji token type. */
	EMOJI = 7
)

/* String token types that correspond to token type int constants */
var TOKEN_TYPES = []string{
	"<ALPHANUM>",
	"<NUM>",
	"<SOUTHEAST_ASIAN>",
	"<IDEOGRAPHIC>",
	"<HIRAGANA>",
	"<KATAKANA>",
	"<HANGUL>",
	"<EMOJI>",
}

/** A grammar-based tokenizer constructed with JFlex.
 * <p>
 * This class implements the Word Break rules from the
 * Unicode Text Segmentation algorithm, as specified in
 * <a href="http://unicode.org/reports/tr29/">Unicode Standard Annex #29</a>.
 * <p>Many applications have specific tokenizer needs.  If this tokenizer does
 * not suit your application, please consider copying this source code
 * directory to your project and maintaining your own grammar-based tokenizer.
 */

type StandardTokenizer struct {
	*Tokenizer

	// A private instance of the JFlex-constructed scanner
	scanner StandardTokenizerInterface

	skippedPositions int
	maxTokenLength   int

	// this tokenizer generates three attributes:
	// term offset, positionIncrement and type

	termAtt    CharTermAttribute
	offsetAtt  OffsetAttribute
	posIncrAtt PositionIncrementAttribute
	typeAtt    TypeAttribute
}

/**
 * Creates a new instance of the {@link org.apache.lucene.analysis.standard.StandardTokenizer}.  Attaches
 * the <code>input</code> to the newly created JFlex scanner.

 * See http://issues.apache.org/jira/browse/LUCENE-1068
 */
func NewStandardTokenizer(matchVersion util.Version, input io.RuneReader) *StandardTokenizer {
	ans := &StandardTokenizer{
		Tokenizer:      NewTokenizer(input),
		maxTokenLength: DEFAULT_MAX_TOKEN_LENGTH,
	}
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	ans.offsetAtt = ans.Attributes().Add("OffsetAttribute").(OffsetAttribute)
	ans.posIncrAtt = ans.Attributes().Add("PositionIncrementAttribute").(PositionIncrementAttribute)
	ans.typeAtt = ans.Attributes().Add("TypeAttribute").(TypeAttribute)
	ans.init(matchVersion)
	return ans
}

func (t *StandardTokenizer) init(matchVersion util.Version) {
	// GoLucene support >=4.5 only
	t.scanner = newStandardTokenizerImpl(nil)
}

func (t *StandardTokenizer) IncrementToken() (bool, error) {
	t.Attributes().Clear()
	t.skippedPositions = 0

	for {
		tokenType, err := t.scanner.nextToken()
		if tokenType == YYEOF || err != nil {
			return false, err
		}

		if t.scanner.yylength() <= t.maxTokenLength {
			t.posIncrAtt.SetPositionIncrement(t.skippedPositions + 1)
			t.scanner.text(t.termAtt)
			start := t.scanner.yychar()
			t.offsetAtt.SetOffset(t.CorrectOffset(start), t.CorrectOffset(start+t.termAtt.Length()))
			t.typeAtt.SetType(TOKEN_TYPES[tokenType])
			return true, nil
		} else {
			// When we skip a too-long term, we still increment the
			// position increment
			t.skippedPositions++
		}
	}

}

func (t *StandardTokenizer) End() error {
	err := t.Tokenizer.End()
	if err == nil {
		// set final offset
		finalOffset := t.CorrectOffset(t.scanner.yychar() + t.scanner.yylength())
		t.offsetAtt.SetOffset(finalOffset, finalOffset)
		// adjust any skipped tokens
		t.posIncrAtt.SetPositionIncrement(t.posIncrAtt.PositionIncrement() + t.skippedPositions)
	}
	return nil
}

func (t *StandardTokenizer) Close() error {
	if err := t.Tokenizer.Close(); err != nil {
		return err
	}
	t.scanner.yyreset(t.Input)
	return nil
}

func (t *StandardTokenizer) Reset() error {
	if err := t.Tokenizer.Reset(); err != nil {
		return err
	}
	t.scanner.yyreset(t.Input)
	t.skippedPositions = 0
	return nil
}

// standard/StandardTokenizerInterface.java

/* This character denotes the end of file */
const YYEOF = -1

/* Internal interface for supporting versioned grammars. */
type StandardTokenizerInterface interface {
	// Copies the matched text into the CharTermAttribute
	text(CharTermAttribute)
	// Returns the current position.
	yychar() int
	// Resets the scanner to read from a new input stream.
	// Does not close the old reader.
	//
	// All internal variables are reset, the old input stream cannot be
	// reused (internal buffer) is discarded and lost). Lexical state
	// is set to ZZ_INITIAL.
	yyreset(io.RuneReader)
	// Returns the length of the matched text region.
	yylength() int
	// Resumes scanning until the next regular expression is matched,
	// the end of input is encountered or an I/O-Error occurs.
	nextToken() (int, error)
}
