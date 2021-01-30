package classic

import (
// "fmt"
)

var jjbitVec0 = []int64{1, 0, 0, 0}
var jjbitVec1 = []uint64{0xfffffffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}
var jjbitVec3 = []uint64{0x0, 0x0, 0xffffffffffffffff, 0xffffffffffffffff}
var jjbitVec4 = []uint64{0xfffefffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}

var jjnextStates = []int{
	37, 39, 40, 17, 18, 20, 42, 45, 31, 46, 43, 22, 23, 25, 26, 24,
	25, 26, 45, 31, 46, 44, 47, 35, 22, 28, 29, 27, 27, 30, 30, 0,
	1, 2, 4, 5,
}

var jjstrLiteralImages = map[int]string{
	0: "", 11: "\u0053", 12: "\055",
	14: "\050", 15: "\051", 16: "\072", 17: "\052", 18: "\136",
	25: "\133", 26: "\173", 28: "\124\117", 29: "\135", 30: "\175",
}

var jjnewLexState = []int{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 0, -1, -1, -1, -1, -1, -1,
	1, 1, 2, -1, 2, 2, -1, -1,
}

var jjtoToken = []int64{0x1ffffff01}

type TokenManager struct {
	curLexState     int
	defaultLexState int
	jjnewStateCnt   int
	jjround         int
	jjmatchedPos    int
	jjmatchedKind   int

	input_stream CharStream
	jjrounds     []int
	jjstateSet   []int
	curChar      rune
}

func newTokenManager(stream CharStream) *TokenManager {
	return &TokenManager{
		curLexState:     2,
		defaultLexState: 2,
		input_stream:    stream,
		jjrounds:        make([]int, 49),
		jjstateSet:      make([]int, 98),
	}
}

// L41
func (tm *TokenManager) jjStopAtPos(pos, kind int) int {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	return pos + 1
}

func (tm *TokenManager) jjMoveStringLiteralDfa0_2() int {
	switch tm.curChar {
	case 40:
		return tm.jjStopAtPos(0, 14)
	case 41:
		return tm.jjStopAtPos(0, 15)
	case 42:
		return tm.jjStartNfaWithStates_2(0, 17, 49)
	case 43:
		return tm.jjStartNfaWithStates_2(0, 11, 15)
	case 45:
		return tm.jjStartNfaWithStates_2(0, 12, 15)
	case 58:
		return tm.jjStopAtPos(0, 16)
	case 91:
		return tm.jjStopAtPos(0, 25)
	case 94:
		return tm.jjStopAtPos(0, 18)
	case 123:
		return tm.jjStopAtPos(0, 26)
	default:
		return tm.jjMoveNfa_2(0, 0)
	}
}

func (tm *TokenManager) jjStartNfaWithStates_2(pos, kind, state int) int {
	var err error
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	tm.curChar, err = tm.input_stream.readChar()
	if err != nil {
		return pos + 1
	}

	return tm.jjMoveNfa_2(state, pos+1)
}

// L87

func (tm *TokenManager) jjMoveNfa_2(startState, curPos int) int {
	startsAt := 0
	tm.jjnewStateCnt = 49
	i := 1
	tm.jjstateSet[0] = startState
	kind := 0x7fffffff
	for {
		if tm.jjround++; tm.jjround == 0x7fffffff {
			tm.reInitRounds()
		}
		if tm.curChar < 64 {
			l := int64(1 << uint(tm.curChar))
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49, 33:
					if (0xfbff7cf8ffffd9ff & uint64(l)) == 0 {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 0:
					if (0xfbff54f8ffffd9ff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					} else if (0x100002600 & l) != 0 {
						if kind > 7 {
							kind = 7
						}
					} else if (0x280200000000 & l) != 0 {
						tm.jjstateSet[tm.jjnewStateCnt] = 15
						tm.jjnewStateCnt++
					} else if tm.curChar == 47 {
						tm.jjCheckNAddStates(0, 2)
					} else if tm.curChar == 34 {
						tm.jjCheckNAddStates(3, 5)
					}
					if (0x7bff50f8ffffd9ff & l) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					} else if tm.curChar == 42 {
						if kind > 22 {
							kind = 22
						}
					} else if tm.curChar == 33 {
						if kind > 10 {
							kind = 10
						}
					}
					if tm.curChar == 38 {
						tm.jjstateSet[tm.jjnewStateCnt] = 4
						tm.jjnewStateCnt++
					}
					break
				case 4:
					if tm.curChar == 38 && kind > 8 {
						kind = 8
					}
					break
				case 5:
					if tm.curChar == 38 {
						tm.jjstateSet[tm.jjnewStateCnt] = 4
						tm.jjnewStateCnt++
					}
					break
				case 13:
					if tm.curChar == 33 && kind > 10 {
						kind = 10
					}
					break
				case 14:
					if (0x280200000000 & l) != 0 {
						tm.jjstateSet[tm.jjnewStateCnt] = 15
						tm.jjnewStateCnt++
					}
					break
				case 15:
					if (0x100002600&l) != 0 && kind > 13 {
						kind = 13
					}
					break
				case 16:
					if tm.curChar == 34 {
						tm.jjCheckNAddStates(3, 5)
					}
					break
				case 17:
					if (0xfffffffbffffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(3, 5)
					}
					break
				case 19:
					tm.jjCheckNAddStates(3, 5)
					break
				case 20:
					if tm.curChar == 34 && kind > 19 {
						kind = 19
					}
					break
				case 22:
					if (0x3ff000000000000 & l) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddStates(11, 14)
					break
				case 23:
					if tm.curChar == 46 {
						tm.jjCheckNAdd(24)
					}
					break
				case 24:
					if (0x3ff000000000000 & l) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddStates(15, 17)
					break
				case 25:
					if (0x7bff78f8ffffd9ff & l) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 27:
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 28:
					if (0x7bff78f8ffffd9ff & l) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 30:
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 31:
					if tm.curChar == 42 && kind > 22 {
						kind = 22
					}
					break
				case 32:
					if (0xfbff54f8ffffd9ff & uint64(l)) == 0 {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 35:
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 36, 38:
					if tm.curChar == 47 {
						tm.jjCheckNAddStates(0, 2)
					}
					break
				case 37:
					if (0xffff7fffffffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(0, 2)
					}
					break
				case 40:
					if tm.curChar == 47 && kind > 24 {
						kind = 24
					}
					break
				case 41:
					if (0x7bff50f8ffffd9ff & l) == 0 {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddStates(6, 10)
					break
				case 42:
					if (0x7bff78f8ffffd9ff & l) == 0 {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 44:
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 45:
					if (0x7bff78f8ffffd9ff & l) != 0 {
						tm.jjCheckNAddStates(18, 20)
					}
					break
				case 47:
					tm.jjCheckNAddStates(18, 20)
					break
				}
				if i == startsAt {
					break
				}
			}
		} else if tm.curChar < 128 {
			l := int64(1) << (uint(tm.curChar) & 077)
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					} else if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(35, 35)
					}
					break
				case 0:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					} else if tm.curChar == 92 {
						tm.jjCheckNAddStates(21, 23)
					} else if tm.curChar == 126 {
						if kind > 21 {
							kind = 21
						}
						tm.jjCheckNAddStates(24, 26)
					}
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
					if tm.curChar == 78 {
						tm.jjstateSet[tm.jjnewStateCnt] = 11
						tm.jjnewStateCnt++
					} else if tm.curChar == 124 {
						tm.jjstateSet[tm.jjnewStateCnt] = 8
						tm.jjnewStateCnt++
					} else if tm.curChar == 79 {
						tm.jjstateSet[tm.jjnewStateCnt] = 6
						tm.jjnewStateCnt++
					} else if tm.curChar == 65 {
						tm.jjstateSet[tm.jjnewStateCnt] = 2
						tm.jjnewStateCnt++
					}
					break
				case 1:
					if tm.curChar == 68 && kind > 8 {
						kind = 8
					}
					break
				case 2:
					if tm.curChar == 78 {
						tm.jjstateSet[tm.jjnewStateCnt] = 1
						tm.jjnewStateCnt++
					}
					break
				case 3:
					if tm.curChar == 65 {
						tm.jjstateSet[tm.jjnewStateCnt] = 2
						tm.jjnewStateCnt++
					}
					break
				case 6:
					if tm.curChar == 82 && kind > 9 {
						kind = 9
					}
					break
				case 7:
					if tm.curChar == 79 {
						tm.jjstateSet[tm.jjnewStateCnt] = 6
						tm.jjnewStateCnt++
					}
					break
				case 8:
					if tm.curChar == 124 && kind > 9 {
						kind = 9
					}
					break
				case 9:
					if tm.curChar == 124 {
						tm.jjstateSet[tm.jjnewStateCnt] = 8
						tm.jjnewStateCnt++
					}
					break
				case 10:
					if tm.curChar == 84 && kind > 10 {
						kind = 10
					}
					break
				case 11:
					if tm.curChar == 79 {
						tm.jjstateSet[tm.jjnewStateCnt] = 10
						tm.jjnewStateCnt++
					}
					break
				case 12:
					if tm.curChar == 78 {
						tm.jjstateSet[tm.jjnewStateCnt] = 11
						tm.jjnewStateCnt++
					}
					break
				case 17:
					if (0xffffffffefffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(3, 5)
					}
					break
				case 18:
					if tm.curChar == 92 {
						tm.jjstateSet[tm.jjnewStateCnt] = 19
						tm.jjnewStateCnt++
					}
					break
				case 19:
					tm.jjCheckNAddStates(3, 5)
					break
				case 21:
					if tm.curChar != 126 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddStates(24, 26)
					break
				case 25:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 26:
					if tm.curChar == 92 {
						tm.jjAddStates(27, 28)
					}
					break
				case 27:
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 28:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 29:
					if tm.curChar == 92 {
						tm.jjAddStates(29, 30)
					}
					break
				case 30:
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 32:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 33:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 34:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(35, 35)
					}
					break
				case 35:
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 37:
					tm.jjAddStates(0, 2)
					break
				case 39:
					if tm.curChar == 92 {
						tm.jjstateSet[tm.jjnewStateCnt] = 38
						tm.jjnewStateCnt++
					}
					break
				case 41:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddStates(6, 10)
					break
				case 42:
					if (0x97ffffff87ffffff & uint64(l)) == 0 {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 43:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(44, 44)
					}
					break
				case 44:
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 45:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(18, 20)
					}
					break
				case 46:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(47, 47)
					}
					break
				case 47:
					tm.jjCheckNAddStates(18, 20)
					break
				case 48:
					if tm.curChar == 92 {
						tm.jjCheckNAddStates(21, 23)
					}
					break
				}
				if i <= startsAt { // ==?
					break
				}
			}
		} else {
			hiByte := int(tm.curChar >> 8)
			i1 := hiByte >> 6
			l1 := int64(1 << (uint64(hiByte) & 077))
			i2 := int((tm.curChar & 0xff) >> 6)
			l2 := int64(1 << uint64(tm.curChar&077))
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49, 33:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 0:
					if jjCanMove_0(hiByte, i1, i2, l1, l2) {
						if kind > 7 {
							kind = 7
						}
					}
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					}
					break
				case 15:
					if jjCanMove_0(hiByte, i1, i2, l1, l2) && kind > 13 {
						kind = 13
					}
					break
				case 17, 19:
					if jjCanMove_1(hiByte, i1, i2, l1, l2) {
						tm.jjCheckNAddStates(3, 5)
					}
					break
				case 25:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 27:
					if !jjCanMove_1(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(25, 26)
					break
				case 28:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 30:
					if !jjCanMove_1(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 21 {
						kind = 21
					}
					tm.jjCheckNAddTwoStates(28, 29)
					break
				case 32:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 35:
					if !jjCanMove_1(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 23 {
						kind = 23
					}
					tm.jjCheckNAddTwoStates(33, 34)
					break
				case 37:
					if jjCanMove_1(hiByte, i1, i2, l1, l2) {
						tm.jjAddStates(0, 2)
					}
					break
				case 41:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddStates(6, 10)
					break
				case 42:
					if !jjCanMove_2(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 44:
					if !jjCanMove_1(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 20 {
						kind = 20
					}
					tm.jjCheckNAddTwoStates(42, 43)
					break
				case 45:
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						tm.jjCheckNAddStates(18, 20)
					}
					break
				case 47:
					if jjCanMove_1(hiByte, i1, i2, l1, l2) {
						tm.jjCheckNAddStates(18, 20)
					}
					break
				}
				if i == startsAt {
					break
				}
			}
		}
		if kind != 0x7fffffff {
			tm.jjmatchedKind = kind
			tm.jjmatchedPos = curPos
			kind = 0x7fffffff
		}
		curPos++
		i = tm.jjnewStateCnt
		tm.jjnewStateCnt = startsAt
		startsAt = 49 - tm.jjnewStateCnt
		if i == startsAt {
			return curPos
		}
		var err error
		if tm.curChar, err = tm.input_stream.readChar(); err != nil {
			return curPos
		}
	}
	panic("should not be here")
}

func jjCanMove_0(hiByte, i1, i2 int, l1, l2 int64) bool {
	switch hiByte {
	case 48:
		return (jjbitVec0[i2] & 12) != 0
	}
	return false
}

func jjCanMove_1(hiByte, i1, i2 int, l1, l2 int64) bool {
	switch hiByte {
	case 0:
		return ((jjbitVec3[i2] & uint64(l2)) != 0)
	default:
		if (jjbitVec1[i1] & uint64(l1)) != 0 {
			return true
		}
		return false
	}
}

func jjCanMove_2(hiByte, i1, i2 int, l1, l2 int64) bool {
	switch hiByte {
	case 0:
		return ((jjbitVec3[i2] & uint64(l2)) != 0)
	case 48:
		return ((jjbitVec1[i2] & uint64(l2)) != 0)
	}
	return (jjbitVec4[i1] & uint64(l1)) != 0
}

func (tm *TokenManager) ReInit(stream CharStream) {
	tm.jjmatchedPos = 0
	tm.jjnewStateCnt = 0
	tm.curLexState = tm.defaultLexState
	tm.input_stream = stream
	tm.reInitRounds()
}

func (tm *TokenManager) reInitRounds() {
	tm.jjround = 0x80000001
	for i := 48; i >= 0; i-- {
		tm.jjrounds[i] = 0x80000000
	}
}

// L1027

func (tm *TokenManager) jjFillToken() *Token {
	var curTokenImage string
	if im, ok := jjstrLiteralImages[tm.jjmatchedKind]; ok {
		curTokenImage = im
	} else {
		curTokenImage = tm.input_stream.image()
	}
	beginLine := tm.input_stream.beginLine()
	beginColumn := tm.input_stream.beginColumn()
	endLine := tm.input_stream.endLine()
	endColumn := tm.input_stream.endColumn()
	t := newToken(tm.jjmatchedKind, curTokenImage)

	t.beginLine = beginLine
	t.endLine = endLine
	t.beginColumn = beginColumn
	t.endColumn = endColumn
	return t
}

func (tm *TokenManager) nextToken() (matchedToken *Token) {
	curPos := 0
	var err error
	var eof = false
EOFLoop:
	for !eof {
		if tm.curChar, err = tm.input_stream.beginToken(); err != nil {
			tm.jjmatchedKind = 0
			matchedToken = tm.jjFillToken()
			return
		}

		switch tm.curLexState {
		case 0:
			tm.jjmatchedKind = 0x7fffffff
			tm.jjmatchedPos = 0
			curPos = tm.jjMoveStringLiteralDfa0_0()
			break
		case 1:
			tm.jjmatchedKind = 0x7fffffff
			tm.jjmatchedPos = 0
			curPos = tm.jjMoveStringLiteralDfa0_1()
			break
		case 2:
			tm.jjmatchedKind = 0x7fffffff
			tm.jjmatchedPos = 0
			curPos = tm.jjMoveStringLiteralDfa0_2()
		}

		if tm.jjmatchedKind != 0x7fffffff {
			if tm.jjmatchedPos+1 < curPos {
				tm.input_stream.backup(curPos - tm.jjmatchedPos - 1)
			}
			if (jjtoToken[tm.jjmatchedKind>>6] & (int64(1) << uint(tm.jjmatchedKind&077))) != 0 {
				matchedToken = tm.jjFillToken()
				if jjnewLexState[tm.jjmatchedKind] != -1 {
					tm.curLexState = jjnewLexState[tm.jjmatchedKind]
					return matchedToken
				}
				return matchedToken
			} else {
				if n := jjnewLexState[tm.jjmatchedKind]; n != -1 {
					tm.curLexState = jjnewLexState[tm.jjmatchedKind]
					continue EOFLoop
				}
				continue
			}
		}
		error_line := tm.input_stream.endLine()
		error_column := tm.input_stream.endColumn()
		var error_after string
		var eofSeen = false
		if _, err = tm.input_stream.readChar(); err == nil {
			tm.input_stream.backup(1)
			if curPos > 1 {
				error_after = tm.input_stream.image()
			}
		} else {
			eofSeen = true
			if curPos > 1 {
				error_after = tm.input_stream.image()
			}
			if tm.curChar == '\n' || tm.curChar == '\r' {
				error_line++
				error_column = 0
			} else {
				error_column++
			}
		}

		panic(newTokenMgrError(eofSeen, tm.curLexState, error_line,
			error_column, error_after, tm.curChar, LEXICAL_ERROR))
	}
	panic("should not be here")
}

func (tm *TokenManager) jjMoveStringLiteralDfa0_0() int {
	return tm.jjMoveNfa_0(0, 0)
}

func (tm *TokenManager) jjMoveNfa_0(startState, curPos int) int {
	startsAt := 0
	tm.jjnewStateCnt = 3
	i := 1
	tm.jjstateSet[0] = startState
	kind := 0x7fffffff
	for {
		tm.jjround++
		if tm.jjround == 0x7fffffff {
			tm.reInitRounds()
		}
		if tm.curChar < 64 {
			l := 1 << tm.curChar
			for {
				i--
				switch tm.jjstateSet[i] {
				case 0:
					if (0x3ff000000000000 & l) == 0 {
						break
					}
					if kind > 27 {
						kind = 27
					}
					tm.jjAddStates(31, 32)
					break
				case 1:
					if tm.curChar == 46 {
						tm.jjCheckNAdd(2)
					}
					break
				case 2:
					if (0x3ff000000000000 & l) == 0 {
						break
					}
					if kind > 27 {
						kind = 27
					}
					tm.jjCheckNAdd(2)
					break
				}
				if i == startsAt {
					break
				}
			}
		} else if tm.curChar < 128 {
			// l := 1 << (tm.curChar & 077)
			for {
				i--
				switch tm.jjstateSet[i] {
				default:
					break
				}
				if i == startsAt {
					break
				}
			}
		} else {
			// hiByte := int(tm.curChar >> 8)
			// i1 := hiByte >> 6
			// l1 := uint64(1) << (hiByte & 077)
			// i2 := (tm.curChar & 0xff) >> 6
			// l2 := uint64(1) << (tm.curChar & 077)
			for {
				i--
				switch tm.jjstateSet[i] {
				default:
					break
				}
				if i == startsAt {
					break
				}
			}
		}
		if kind != 0x7fffffff {
			tm.jjmatchedKind = kind
			tm.jjmatchedPos = curPos
			kind = 0x7fffffff
		}
		curPos++
		i = tm.jjnewStateCnt
		tm.jjnewStateCnt = startsAt
		startsAt = 3 - startsAt
		if i == startsAt {
			return curPos
		}

		var err error
		if tm.curChar, err = tm.input_stream.readChar(); err != nil {
			return curPos
		}
	}
}

// L1137
func (tm *TokenManager) jjCheckNAdd(state int) {
	if tm.jjrounds[state] != tm.jjround {
		tm.jjstateSet[tm.jjnewStateCnt] = state
		tm.jjnewStateCnt++
		tm.jjrounds[state] = tm.jjround
	}
}
func (tm *TokenManager) jjAddStates(start, end int) {
	for {
		tm.jjstateSet[tm.jjnewStateCnt] = jjnextStates[start]
		tm.jjnewStateCnt++
		start++
		if start >= end {
			break
		}
	}
}

// L1151

func (tm *TokenManager) jjCheckNAddTwoStates(state1, state2 int) {
	tm.jjCheckNAdd(state1)
	tm.jjCheckNAdd(state2)
}

func (tm *TokenManager) jjCheckNAddStates(start, end int) {
	assert(start < end)
	assert(start >= 0)
	assert(end <= len(jjnextStates))
	for {
		tm.jjCheckNAdd(jjnextStates[start])
		start++
		if start >= end {
			break
		}
	}
}

func (tm *TokenManager) jjStopStringLiteralDfa_1(pos int, active0 int64) int {
	switch pos {
	case 0:
		if (active0 & 0x10000000) != 0 {
			tm.jjmatchedKind = 32
			return 6
		}
		return -1
	default:
		return -1
	}
}

func (tm *TokenManager) jjMoveStringLiteralDfa1_1(active0 int64) int {
	var err error

	tm.curChar, err = tm.input_stream.readChar()
	if err != nil {
		tm.jjStopStringLiteralDfa_1(0, active0)
		return 1
	}

	switch tm.curChar {
	case 79:
		if (active0 & 0x10000000) != 0 {
			return tm.jjStartNfaWithStates_1(1, 28, 6)
		}
		break
	default:
		break
	}
	return tm.jjStartNfa_1(0, active0)
}

func (tm *TokenManager) jjStartNfaWithStates_1(pos, kind, state int) int {
	tm.jjmatchedKind = kind
	tm.jjmatchedPos = pos
	var err error
	tm.curChar, err = tm.input_stream.readChar()
	if err != nil {
		return pos + 1
	}

	return tm.jjMoveNfa_1(state, pos+1)
}

func (tm *TokenManager) jjStartNfa_1(pos int, active0 int64) int {
	return tm.jjMoveNfa_1(tm.jjStopStringLiteralDfa_1(pos, active0), pos+1)
}

func (tm *TokenManager) jjMoveStringLiteralDfa0_1() int {
	switch tm.curChar {
	case 84:
		return tm.jjMoveStringLiteralDfa1_1(0x10000000)
	case 93:
		return tm.jjStopAtPos(0, 29)
	case 125:
		return tm.jjStopAtPos(0, 30)
	default:
		return tm.jjMoveNfa_1(0, 0)
	}
}

func (tm *TokenManager) jjMoveNfa_1(startState, curPos int) int {
	startsAt := 0
	tm.jjnewStateCnt = 7
	i := 1
	tm.jjstateSet[0] = startState
	kind := 0x7fffffff
	for {
		tm.jjround++
		if tm.jjround == 0x7fffffff {
			tm.reInitRounds()
		}
		if tm.curChar < 64 {
			l := int64(1 << tm.curChar)
			for {
				i--
				switch tm.jjstateSet[i] {
				case 0:
					if (0xfffffffeffffffff & uint64(l)) != 0 {
						if kind > 32 {
							kind = 32
						}
						tm.jjCheckNAdd(6)
					}
					if (0x100002600 & l) != 0 {
						if kind > 7 {
							kind = 7
						}
					} else if tm.curChar == 34 {
						tm.jjCheckNAddTwoStates(2, 4)
					}
					break
				case 1:
					if tm.curChar == 34 {
						tm.jjCheckNAddTwoStates(2, 4)
					}
					break
				case 2:
					if (0xfffffffbffffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(33, 35)
					}
					break
				case 3:
					if tm.curChar == 34 {
						tm.jjCheckNAddStates(33, 35)
					}
					break
				case 5:
					if tm.curChar == 34 && kind > 31 {
						kind = 31
					}
					break
				case 6:
					if (0xfffffffeffffffff & uint64(l)) == 0 {
						break
					}
					if kind > 32 {
						kind = 32
					}
					tm.jjCheckNAdd(6)
					break
				default:
					break
				}
				if i == startsAt {
					break
				}
			}
		} else if tm.curChar < 128 {
			l := 1 << (tm.curChar & 077)
			for {
				i--
				switch tm.jjstateSet[i] {
				case 0:
				case 6:
					if (0xdfffffffdfffffff & uint64(l)) == 0 {
						break
					}
					if kind > 32 {
						kind = 32
					}
					tm.jjCheckNAdd(6)
					break
				case 2:
					tm.jjAddStates(33, 35)
					break
				case 4:
					if tm.curChar == 92 {
						tm.jjstateSet[tm.jjnewStateCnt] = 3
						tm.jjnewStateCnt++
					}
					break
				}

				if i == startsAt {
					break
				}
			}
		} else {
			hiByte := int(tm.curChar >> 8)
			i1 := hiByte >> 6
			l1 := int64(1 << (hiByte & 077))
			i2 := int((tm.curChar & 0xff) >> 6)
			l2 := int64(1 << (tm.curChar & 077))
			for {
				i--
				switch tm.jjstateSet[i] {
				case 0:
					if jjCanMove_0(hiByte, i1, i2, l1, l2) {
						if kind > 7 {
							kind = 7
						}
					}

					if jjCanMove_1(hiByte, i1, i2, l1, l2) {
						if kind > 32 {
							kind = 32
						}
						tm.jjCheckNAdd(6)
					}
					break
				case 2:
					if jjCanMove_1(hiByte, i1, i2, l1, l2) {
						tm.jjAddStates(33, 35)
					}
					break
				case 6:
					if !jjCanMove_1(hiByte, i1, i2, l1, l2) {
						break
					}
					if kind > 32 {
						kind = 32
					}
					tm.jjCheckNAdd(6)
					break
				default:
					break
				}
				if i == startsAt {
					break
				}
			}
		}
		if kind != 0x7fffffff {
			tm.jjmatchedKind = kind
			tm.jjmatchedPos = curPos
			kind = 0x7fffffff
		}
		curPos++
		i = tm.jjnewStateCnt
		tm.jjnewStateCnt = startsAt
		startsAt = 7 - startsAt
		if i == startsAt {
			return curPos
		}
		var err error
		tm.curChar, err = tm.input_stream.readChar()
		if err != nil {
			return curPos
		}
	}
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
