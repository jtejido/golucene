package fst

import (
	"fmt"
	"github.com/jtejido/golucene/core/util"
	"reflect"
)

type FSTStore interface {
	util.Accountable
	Init(in util.DataInput, numBytes int64) error
	Size() int64
	ReverseBytesReader() BytesReader
	WriteTo(out util.DataOutput) error
}

type ByteHeap []byte

func (h ByteHeap) Len() int           { return len(h) }
func (h ByteHeap) Less(i, j int) bool { return int(h[i]) < int(h[j]) }
func (h ByteHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *ByteHeap) Push(x interface{}) {
	*h = append(*h, x.(byte))
}

func (h *ByteHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type OnHeapFSTStore struct {
	bytes        *BytesStore
	bytesArray   []byte
	maxBlockBits int
}

func newOnHeapFSTStore(maxBlockBits int) *OnHeapFSTStore {
	if maxBlockBits < 1 || maxBlockBits > 30 {
		panic(fmt.Sprintf("maxBlockBits should be 1 .. 30; got %d", maxBlockBits))
	}
	return &OnHeapFSTStore{maxBlockBits: maxBlockBits}
}

func (onfsts *OnHeapFSTStore) Init(in util.DataInput, numBytes int64) (err error) {
	if numBytes > 1<<uint(onfsts.maxBlockBits) {
		// FST is big: we need multiple pages
		onfsts.bytes, err = newBytesStoreFromInput(in, numBytes, 1<<uint(onfsts.maxBlockBits))
	} else {
		// FST fits into a single block: use ByteArrayBytesStoreReader for less overhead
		onfsts.bytesArray = make([]byte, numBytes)
		err = in.ReadBytes(onfsts.bytesArray)
	}

	return
}

func (onfsts *OnHeapFSTStore) Size() int64 {
	if onfsts.bytesArray != nil {
		return int64(len(onfsts.bytesArray))
	} else {
		return onfsts.bytes.RamBytesUsed()
	}
}

func (onfsts *OnHeapFSTStore) RamBytesUsed() int64 {
	rbu := util.ShallowSizeOfInstance(reflect.TypeOf(OnHeapFSTStore{}))
	return rbu + onfsts.Size()
}

func (onfsts *OnHeapFSTStore) ReverseBytesReader() BytesReader {
	if onfsts.bytesArray != nil {
		return newReverseBytesReader(onfsts.bytesArray)
	} else {
		return onfsts.bytes.reverseReader()
	}
}

func (onfsts *OnHeapFSTStore) WriteTo(out util.DataOutput) (err error) {
	if onfsts.bytes != nil {
		numBytes := onfsts.bytes.position()
		if err = out.WriteVLong(numBytes); err == nil {
			err = onfsts.bytes.writeTo(out)
		}
	} else {
		assert(onfsts.bytesArray != nil)
		if err = out.WriteVLong(int64(len(onfsts.bytesArray))); err == nil {
			err = out.WriteBytes(onfsts.bytesArray)
		}
	}

	return
}
