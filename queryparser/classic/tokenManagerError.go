package classic

import (
	"fmt"
	"strconv"
)

const (
	LEXICAL_ERROR = iota
	STATIC_LEXER_ERROR
	INVALID_LEXICAL_STATE
	LOOP_DETECTED
)

type TokenManagerError struct {
}

func newTokenMgrError(eofSeen bool, lexState, errorLine, errorColumn int,
	errorAfter string, curChar rune, reason int) *TokenManagerError {
	panic(LexicalError(eofSeen, lexState, errorLine, errorColumn, errorAfter, curChar))
}

func (err *TokenManagerError) Error() string {
	panic("not implemented yet")
}

func LexicalError(EOFSeen bool, lexState, errorLine, errorColumn int, errorAfter string, curChar rune) string {
	var s string
	if EOFSeen {
		s += "<EOF> "
	} else {
		s += "\""
		s += addEscapes(string(curChar))
		s += "\""
		s += fmt.Sprintf("(%d), ", int(curChar))
	}
	return fmt.Sprintf("Lexical error at line %d, column %d.  Encountered: %safter : \"%s\"", errorLine, errorColumn, s, addEscapes(errorAfter))
}

func addEscapes(str string) string {
	var s string
	for i := 0; i < len(str); i++ {
		switch rune(str[i]) {
		case 0:
			continue
		case '\b':
			s += "\\b"
			continue
		case '\t':
			s += "\\t"
			continue
		case '\n':
			s += "\\n"
			continue
		case '\f':
			s += "\\f"
			continue
		case '\r':
			s += "\\r"
			continue
		case '"':
			s += "\\\""
			continue
		case '\'':
			s += "\\'"
			continue
		case '\\':
			s += "\\\\"
			continue
		default:
			ch := int(rune(str[i]))
			if ch < 0x20 || ch > 0x7e {
				ss := fmt.Sprintf("0000%s", strconv.FormatInt(int64(ch), 16))
				s += fmt.Sprintf("\\u%s", ss[len(s)-4:len(s)])
			} else {
				s += string(ch)
			}
			continue
		}
	}
	return s
}
