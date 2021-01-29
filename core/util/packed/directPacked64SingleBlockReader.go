package packed

import (
	"github.com/jtejido/golucene/core/store"
)

type DirectPacked64SingleBlockReader struct {
	*ReaderImpl
	*abstractReader
	in             store.IndexInput
	bitsPerValue   uint32
	startPointer   int64
	valuesPerBlock uint32
	mask           int64
}

func newDirectPacked64SingleBlockReader(bitsPerValue uint32, valueCount int, in store.IndexInput) (*DirectPacked64SingleBlockReader, error) {
	ans := &DirectPacked64SingleBlockReader{
		ReaderImpl:     newReaderImpl(valueCount),
		in:             in,
		bitsPerValue:   bitsPerValue,
		startPointer:   in.FilePointer(),
		valuesPerBlock: 64 / bitsPerValue,
		mask:           ^(^0 << bitsPerValue),
	}
	ans.abstractReader = newReader(ans)
	return ans, nil
}

func (a *DirectPacked64SingleBlockReader) Get(index int) int64 {
	blockOffset := int64(index) / int64(a.valuesPerBlock)
	skip := (blockOffset) << 3
	var err error
	var block int64

	if err = a.in.Seek(a.startPointer + skip); err != nil {
		panic(err.Error())
	}

	if block, err = a.in.ReadLong(); err != nil {
		panic(err.Error())
	}
	offsetInBlock := index % int(a.valuesPerBlock)
	return (block >> (int64(offsetInBlock) * int64(a.bitsPerValue))) & a.mask
}

func (a *DirectPacked64SingleBlockReader) RamBytesUsed() int64 {
	return 0
}
