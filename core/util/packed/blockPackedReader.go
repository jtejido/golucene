package packed

import (
	"fmt"
	"github.com/jtejido/golucene/core/store"
	"github.com/jtejido/golucene/core/util"
)

type BlockPackedReader interface {
	util.Accountable
	Get(index int64) int64
}

type BlockPackedReaderImpl struct {
	blockShift, blockMask int
	valueCount            int64
	minValues             []int64
	subReaders            []PackedIntsReader
}

func NewBlockPackedReader(in store.IndexInput, packedIntsVersion int32, blockSize int, valueCount int64, direct bool) (*BlockPackedReaderImpl, error) {
	a := new(BlockPackedReaderImpl)
	a.valueCount = valueCount
	a.blockShift = checkBlockSize(blockSize, BPW_MIN_BLOCK_SIZE, BPW_MAX_BLOCK_SIZE)
	a.blockMask = blockSize - 1
	numBlocks := numBlocks(valueCount, blockSize)
	a.subReaders = make([]PackedIntsReader, numBlocks)
	var err error
	for i := 0; i < numBlocks; i++ {
		var b byte
		if b, err = in.ReadByte(); err != nil {
			return nil, err
		}
		token := uint32(b) & 0xFF
		bitsPerValue := token >> BPV_SHIFT
		if bitsPerValue > 64 {
			return nil, fmt.Errorf("Corrupted")
		}
		if (token & MIN_VALUE_EQUALS_0) == 0 {
			if a.minValues == nil {
				a.minValues = make([]int64, numBlocks)
			}
			var t int64
			if t, err = readVLong(in); err != nil {
				return nil, err
			}
			a.minValues[i] = util.ZigZagEncodeLong(1 + t)
		}
		if bitsPerValue == 0 {
			a.subReaders[i] = newNilReader(blockSize)
		} else {
			size := int32(valueCount - int64(i*blockSize))
			if size > int32(blockSize) {
				size = int32(blockSize)
			}

			if direct {
				pointer := in.FilePointer()
				if a.subReaders[i], err = DirectReaderNoHeader(in, PackedFormat(PACKED), packedIntsVersion, size, bitsPerValue); err != nil {
					return nil, err
				}
				if err = in.Seek(pointer + PackedFormat(PACKED).ByteCount(packedIntsVersion, size, bitsPerValue)); err != nil {
					return nil, err
				}
			} else {
				if a.subReaders[i], err = ReaderNoHeader(in, PackedFormat(PACKED), packedIntsVersion, size, bitsPerValue); err != nil {
					return nil, err
				}
			}
		}
	}

	return a, nil
}

func (a *BlockPackedReaderImpl) Get(index int64) int64 {
	assert(index >= 0 && index < a.valueCount)
	block := (index >> a.blockShift)
	idx := int(index & int64(a.blockMask))

	var t int64
	if a.minValues != nil {
		t += a.minValues[block] + a.subReaders[block].Get(idx)
	}

	return t
}

func (a *BlockPackedReaderImpl) RamBytesUsed() int64 {
	var size int64
	for _, reader := range a.subReaders {
		size += reader.RamBytesUsed()
	}
	return size
}
