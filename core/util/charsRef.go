package util

// util/BytesRef.java

/* An empty byte slice for convenience */
var EMPTY_CHARS = []rune{}

/**
 * Represents char[], as a slice (offset + length) into an existing char[].
 * The {@link #chars} member should never be null; use
 * {@link #EMPTY_CHARS} if necessary.
 * @lucene.internal
 */
type CharsRef struct {
	// The contents of the BytesRef.
	Chars  []rune
	Offset int
	Length int
}

func NewEmptyCharsRef() *CharsRef {
	return NewCharsRefFrom(EMPTY_CHARS)
}

func NewCharsRef(chars []rune, offset, length int) *CharsRef {
	return &CharsRef{
		Chars:  chars,
		Offset: offset,
		Length: length,
	}
}

func NewCharsRefFrom(chars []rune) *CharsRef {
	return NewCharsRef(chars, 0, len(chars))
}

/*
Expert: compares the byte against another BytesRef, returning true if
the bytes are equal.
*/
func (br *CharsRef) charsEquals(other []rune) bool {
	if br.Length != len(other) {
		return false
	}
	for i, v := range br.ToChars() {
		if v != other[i] {
			return false
		}
	}
	return true
}

func (br *CharsRef) String() string {
	panic("not implemented yet")
}

func (br *CharsRef) ToChars() []rune {
	return br.Chars[br.Offset : br.Offset+br.Length]
}

func (a *CharsRef) copyChars(other *CharsRef) {
	if len(a.Chars)-a.Offset < other.Length {
		a.Chars = make([]rune, other.Length)
		a.Offset = 0
	}
	copy(a.Chars, other.Chars[other.Offset:other.Offset+other.Length])
	a.Length = other.Length
}

func (a *CharsRef) CharAt(index int) rune {
	// NOTE: must do a real check here to meet the specs of CharSequence
	if index < 0 || index >= a.Length {
		panic("index out of bounds")
	}

	return a.Chars[a.Offset+index]
}

type CharsRefs []*CharsRef

func (br CharsRefs) Len() int {
	return len(br)
}

func (br CharsRefs) Less(i, j int) bool {
	aChars, bChars := br[i], br[j]
	aLen, bLen := aChars.Length, bChars.Length

	for i, v := range aChars.Chars {
		if i >= bLen {
			break
		}
		if int(v) < int(bChars.Chars[i]) {
			return true
		} else if int(v) > int(bChars.Chars[i]) {
			return false
		}
	}

	// One is a prefix of the other, or, they are equal:
	return aLen < bLen
}

func (br CharsRefs) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}

// util/CharsRefBuilderBuilder.java

type CharsRefBuilder struct {
	ref *CharsRef
}

func NewCharsRefBuilder() *CharsRefBuilder {
	return &CharsRefBuilder{
		ref: NewEmptyCharsRef(),
	}
}

/* Return a reference to the bytes of this build. */
func (b *CharsRefBuilder) Chars() []rune {
	return b.ref.Chars
}

/* Return the number of bytes in this buffer. */
func (b *CharsRefBuilder) Length() int {
	return b.ref.Length
}

/* Set the length. */
func (b *CharsRefBuilder) SetLength(length int) {
	b.ref.Length = length
}

/* Return the char at the given offset. */
func (b *CharsRefBuilder) At(offset int) rune {
	return b.ref.Chars[offset]
}

/* Set a byte. */
func (b *CharsRefBuilder) Set(offset int, v rune) {
	b.ref.Chars[offset] = v
}

/* Ensure that this builder can hold at least capacity bytes without resizing. */
func (b *CharsRefBuilder) Grow(capacity int) {
	b.ref.Chars = GrowRuneSlice(b.ref.Chars, capacity)
}

func (b *CharsRefBuilder) append(chars []rune) {
	b.Grow(b.ref.Length + len(chars))
	copy(b.ref.Chars[b.ref.Length:], chars)
	b.ref.Length += len(chars)
}

func (b *CharsRefBuilder) clear() {
	b.SetLength(0)
}

func (b *CharsRefBuilder) Copy(ref []rune) {
	b.clear()
	b.append(ref)
}

func (b *CharsRefBuilder) Get() *CharsRef {
	assert2(b.ref.Offset == 0, "Modifying the offset of the returned ref is illegal")
	return b.ref
}
