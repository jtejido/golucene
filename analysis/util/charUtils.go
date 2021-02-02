package util

import (
	"github.com/jtejido/golucene/core/util"
	"unicode"
)

// util/CharacterUtils.java

type CharacterUtilsSPI interface{}

/*
Characterutils provides a unified interface to Character-related
operations to implement backwards compatible character operations
based on a version instance.
*/
type CharacterUtils struct {
	CharacterUtilsSPI
}

/* Returns a Characters implementation according to the given Version instance. */
func GetCharacterUtils(matchVersion util.Version) *CharacterUtils {
	return &CharacterUtils{}
}

/* Converts each unicode codepoint to lowerCase via unicode.ToLower(). */
func (cu *CharacterUtils) ToLowerCase(buffer []rune) {
	for i, v := range buffer {
		buffer[i] = unicode.ToLower(v)
	}
}

/* Converts each unicode codepoint to lowerCase via unicode.ToUpper(). */
func (cu *CharacterUtils) ToUpperCase(buffer []rune) {
	for i, v := range buffer {
		buffer[i] = unicode.ToUpper(v)
	}
}

const (
	MIN_SUPPLEMENTARY_CODE_POINT = 0x010000
	MIN_HIGH_SURROGATE           = 0xD800
	MIN_LOW_SURROGATE            = 0xDC00
	MAX_LOW_SURROGATE            = 0xDFFF
	MAX_HIGH_SURROGATE           = 0xDBFF
)

func IsHighSurrogate(r rune) bool {
	return r >= MIN_HIGH_SURROGATE && r < (MAX_HIGH_SURROGATE+1)
}

func IsLowSurrogate(r rune) bool {
	return r >= MIN_LOW_SURROGATE && r < (MAX_LOW_SURROGATE+1)
}

func ToCodePoint(high, low rune) int {
	// Optimized form of:
	// return ((high - MIN_HIGH_SURROGATE) << 10)
	//         + (low - MIN_LOW_SURROGATE)
	//         + MIN_SUPPLEMENTARY_CODE_POINT;
	return ((int(high) << 10) + int(low)) + (MIN_SUPPLEMENTARY_CODE_POINT - (MIN_HIGH_SURROGATE << 10) - MIN_LOW_SURROGATE)
}

func CodePointAt(a []rune, index, limit int) int {
	c1 := a[index]
	index++
	if IsHighSurrogate(c1) && index < limit {
		c2 := a[index]
		if IsLowSurrogate(c2) {
			return ToCodePoint(c1, c2)
		}
	}
	return int(c1)
}

func CharCount(codePoint int) int {
	if codePoint >= MIN_SUPPLEMENTARY_CODE_POINT {
		return 2
	}

	return 1
}
