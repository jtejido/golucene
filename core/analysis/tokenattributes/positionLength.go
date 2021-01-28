package tokenattributes

import (
	"github.com/jtejido/golucene/core/util"
)

type PositionLengthAttribute interface {
	util.Attribute
	SetPositionLength(int)
}
