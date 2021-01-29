package packed

import (
	"fmt"
	"github.com/jtejido/golucene/core/util"
	"math"
)

const (
	BPW_MIN_BLOCK_SIZE = 64
	BPW_MAX_BLOCK_SIZE = 1 << (30 - 3)
	MIN_VALUE_EQUALS_0 = 1 << 0
	BPV_SHIFT          = 1
)

type iBlockPackedWriter interface {
	flush() error
	add(int64) error
}

type BlockPackedWriter interface {
	Flush() error
	Add(int64) error
	Reset(out util.DataOutput)
	WriteVLong(out util.DataOutput, i int64) error
	Finish() error
	Ord() int64
	WriteValues(bitsRequired uint32) error
}

type basePackedWriterImpl struct {
	spi      iBlockPackedWriter
	out      util.DataOutput
	values   []int64
	blocks   []byte
	off      int
	ord      int64
	finished bool
}

func newBlockPackedWriter(spi iBlockPackedWriter, out util.DataOutput, blockSize int) *basePackedWriterImpl {
	checkBlockSize(blockSize, BPW_MIN_BLOCK_SIZE, BPW_MAX_BLOCK_SIZE)
	return &basePackedWriterImpl{
		spi:    spi,
		values: make([]int64, blockSize),
		out:    out,
	}
}

func (a *basePackedWriterImpl) WriteVLong(out util.DataOutput, i int64) error {
	k := 0
	for (i&^0x7F) != 0 && k < 8 {
		k++
		out.WriteByte(byte((i & 0x7F) | 0x80))
		i >>= 7
	}
	return out.WriteByte(byte(i))
}

func (a *basePackedWriterImpl) Reset(out util.DataOutput) {
	assert(out != nil)
	a.out = out
	a.off = 0
	a.ord = 0
	a.finished = false
}

func (a *basePackedWriterImpl) checkNotFinished() error {
	if a.finished {
		return fmt.Errorf("Already finished")
	}
	return nil
}

func (a *basePackedWriterImpl) Add(l int64) error {
	return a.spi.add(l)
}

func (a *basePackedWriterImpl) Finish() error {
	if err := a.checkNotFinished(); err != nil {
		return err
	}
	if a.off > 0 {
		if err := a.spi.flush(); err != nil {
			return err
		}
	}
	a.finished = true
	return nil
}

func (a *basePackedWriterImpl) Ord() int64 {
	return a.ord
}

func (a *basePackedWriterImpl) Flush() error { return a.spi.flush() }

func (a *basePackedWriterImpl) WriteValues(bitsRequired uint32) error {
	encoder := GetPackedIntsEncoder(PackedFormat(PACKED), VERSION_CURRENT, bitsRequired)
	iterations := len(a.values) / encoder.ByteValueCount()
	blockSize := encoder.ByteBlockCount() * iterations
	if a.blocks == nil || len(a.blocks) < blockSize {
		a.blocks = make([]byte, blockSize)
	}
	if a.off < len(a.values) {
		for i := a.off; i < len(a.values); i++ {
			a.values[i] = 0
		}
	}
	encoder.encodeLongToByte(a.values, a.blocks, iterations)
	blockCount := PackedFormat(PACKED).ByteCount(VERSION_CURRENT, int32(a.off), bitsRequired)
	return a.out.WriteBytes(a.blocks[:blockCount])
}

type BlockPackedWriterImpl struct {
	*basePackedWriterImpl
}

func NewBlockPackedWriter(out util.DataOutput, blockSize int) *BlockPackedWriterImpl {
	owner := new(BlockPackedWriterImpl)
	owner.basePackedWriterImpl = newBlockPackedWriter(owner, out, blockSize)
	return owner
}

func (a *BlockPackedWriterImpl) flush() error {
	assert(a.off > 0)
	min := int64(math.MaxInt64)
	max := int64(math.MinInt64)
	for i := 0; i < a.off; i++ {

		if min < a.values[i] {
			min = a.values[i]
		}
		if max > a.values[i] {
			max = a.values[i]
		}
	}

	delta := max - min
	bitsRequired := 0
	if delta != 0 {
		bitsRequired = UnsignedBitsRequired(delta)
	}
	if bitsRequired == 64 {
		// no need to delta-encode
		min = 0
	} else if min > 0 {
		// make min as small as possible so that writeVLong requires fewer bytes
		min = MaxValue(bitsRequired)
		if min < 0 {
			min = 0
		}
	}

	if min == 0 {
		min = MIN_VALUE_EQUALS_0
	}

	token := (uint(bitsRequired) << BPV_SHIFT) | uint(min)
	if err := a.out.WriteByte(byte(token)); err != nil {
		return err
	}

	if min != 0 {
		if err := a.WriteVLong(a.out, util.ZigZagEncodeLong(min)-1); err != nil {
			return err
		}
	}

	if bitsRequired > 0 {
		if min != 0 {
			for i := 0; i < a.off; i++ {
				a.values[i] -= min
			}
		}
		if err := a.WriteValues(uint32(bitsRequired)); err != nil {
			return err
		}
	}

	a.off = 0
	return nil
}

func (a *BlockPackedWriterImpl) add(l int64) error {
	if err := a.checkNotFinished(); err != nil {
		return err
	}
	if a.off == len(a.values) {
		if err := a.flush(); err != nil {
			return err
		}
	}

	a.values[a.off] = l
	a.off++
	a.ord++
	return nil
}
