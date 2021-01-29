package packed

import (
	"github.com/jtejido/golucene/core/util"
)

func readVLong(in util.DataInput) (int64, error) {
	var b byte
	var err error

	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	if b >= 0 {
		return int64(b), nil
	}

	var i int64
	i = int64(b) & 0x7F
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 7
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 14
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 21
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 28
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 35
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 42
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0x7F) << 49
	if b >= 0 {
		return i, nil
	}
	if b, err = in.ReadByte(); err != nil {
		return 0, err
	}
	i |= (int64(b) & 0xFF) << 56
	return i, nil
}
