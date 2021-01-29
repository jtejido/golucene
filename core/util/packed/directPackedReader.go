package packed

import (
	"fmt"
	"github.com/jtejido/golucene/core/store"
)

type DirectPackedReader struct {
	*ReaderImpl
	*abstractReader
	in                      store.IndexInput
	bitsPerValue            uint32
	startPointer, valueMask int64
}

func newDirectPackedReader(bitsPerValue uint32, valueCount int, in store.IndexInput) (*DirectPackedReader, error) {
	var valueMask int64
	if bitsPerValue == 64 {
		valueMask = -1
	} else {
		valueMask = int64(1<<bitsPerValue) - 1
	}

	ans := &DirectPackedReader{
		ReaderImpl:   newReaderImpl(valueCount),
		in:           in,
		bitsPerValue: bitsPerValue,
		startPointer: in.FilePointer(),
		valueMask:    valueMask,
	}
	ans.abstractReader = newReader(ans)
	return ans, nil
}

func (a *DirectPackedReader) Get(index int) int64 {
	majorBitPos := int64(index * int(a.bitsPerValue))
	elementPos := majorBitPos >> 3

	if err := a.in.Seek(a.startPointer + elementPos); err != nil {
		panic(err.Error())
	}

	bitPos := uint32(majorBitPos & 7)
	// round up bits to a multiple of 8 to find total bytes needed to read
	roundedBits := ((bitPos + a.bitsPerValue + 7) &^ 7)
	// the number of extra bits read at the end to shift out
	shiftRightBits := roundedBits - bitPos - a.bitsPerValue

	var rawValue int64
	var err error
	switch roundedBits >> 3 {
	case 1:
		var b byte
		if b, err = a.in.ReadByte(); err != nil {
			panic(err.Error())
		}

		rawValue = int64(b)
		break
	case 2:
		var t int16
		if t, err = a.in.ReadShort(); err != nil {
			panic(err.Error())
		}

		rawValue = int64(t)
		break
	case 3:
		var t int16
		var b byte
		if t, err = a.in.ReadShort(); err != nil {
			panic(err.Error())
		}

		if b, err = a.in.ReadByte(); err != nil {
			panic(err.Error())
		}

		rawValue = (int64(t) << 8) | (int64(b) & 0xFF)
		break
	case 4:
		var t int32
		if t, err = a.in.ReadInt(); err != nil {
			panic(err.Error())
		}

		rawValue = int64(t)
		break
	case 5:
		var t int32
		var b byte
		if t, err = a.in.ReadInt(); err != nil {
			panic(err.Error())
		}

		if b, err = a.in.ReadByte(); err != nil {
			panic(err.Error())
		}

		rawValue = (int64(t) << 8) | (int64(b) & 0xFF)
		break
	case 6:
		var t int32
		var s int16
		if t, err = a.in.ReadInt(); err != nil {
			panic(err.Error())
		}

		if s, err = a.in.ReadShort(); err != nil {
			panic(err.Error())
		}
		rawValue = (int64(t) << 16) | (int64(s) & 0xFFFF)
		break
	case 7:

		var t int32
		var s int16
		var b byte
		if t, err = a.in.ReadInt(); err != nil {
			panic(err.Error())
		}

		if s, err = a.in.ReadShort(); err != nil {
			panic(err.Error())
		}

		if b, err = a.in.ReadByte(); err != nil {
			panic(err.Error())
		}
		rawValue = (int64(t) << 24) | ((int64(s) & 0xFFFF) << 8) | (int64(b) & 0xFF)
		break
	case 8:
		if rawValue, err = a.in.ReadLong(); err != nil {
			panic(err.Error())
		}
		break
	case 9:
		// We must be very careful not to shift out relevant bits. So we account for right shift
		// we would normally do on return here, and reset it.
		var t int64
		var b byte
		if t, err = a.in.ReadLong(); err != nil {
			panic(err.Error())
		}

		if b, err = a.in.ReadByte(); err != nil {
			panic(err.Error())
		}

		rawValue = (t << (8 - shiftRightBits)) | ((int64(b) & 0xFF) >> shiftRightBits)
		shiftRightBits = 0
		break
	default:
		panic(fmt.Sprintf("bitsPerValue too large: %v", a.bitsPerValue))
	}
	return (rawValue >> shiftRightBits) & a.valueMask
}

func (a *DirectPackedReader) RamBytesUsed() int64 {
	return 0
}
