package tokenattributes

import (
	"fmt"
	"github.com/jtejido/golucene/core/util"
)

type PositionLengthAttribute interface {
	SetPositionLength(int)
	PositionLength() int
}

type PositionLengthAttributeImpl struct {
	positionLength int
}

func newPositionLengthAttributeImpl() util.AttributeImpl {
	return &PositionLengthAttributeImpl{
		positionLength: 1,
	}
}

func (a *PositionLengthAttributeImpl) Interfaces() []string {
	return []string{"PositionLengthAttribute"}
}

func (a *PositionLengthAttributeImpl) Clone() util.AttributeImpl {
	return &PositionLengthAttributeImpl{
		positionLength: a.positionLength,
	}
}

func (a *PositionLengthAttributeImpl) PositionLength() int {
	return a.positionLength
}

func (a *PositionLengthAttributeImpl) SetPositionLength(positionLength int) {
	if a.positionLength < 1 {
		panic(fmt.Sprintf("Position length must be 1 or greater: got %d", a.positionLength))
	}
	a.positionLength = positionLength
}

func (a *PositionLengthAttributeImpl) Clear() {
	a.positionLength = 1
}

func (a *PositionLengthAttributeImpl) CopyTo(target util.AttributeImpl) {
	attr := target.(PositionLengthAttribute)
	attr.SetPositionLength(a.positionLength)
}
