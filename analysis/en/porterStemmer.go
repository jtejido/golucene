package en

import (
	"github.com/jtejido/golucene/core/util"
)

const (
	INITIAL_SIZE = 50
)

type PorterStemmer struct {
	b           []rune
	i, j, k, k0 int
	dirty       bool
}

func newPorterStemmer() *PorterStemmer {
	return &PorterStemmer{
		b: make([]rune, INITIAL_SIZE),
	}
}

func (p *PorterStemmer) reset() { p.i = 0; p.dirty = false }

func (p *PorterStemmer) add(ch rune) {
	if len(p.b) <= p.i {
		p.b = util.GrowRuneSlice(p.b, p.i+1)
	}
	p.b[p.i] = ch
	p.i++
}

func (p *PorterStemmer) String() string { return string(p.b[:p.i]) }

func (p *PorterStemmer) ResultLength() int { return p.i }

func (p *PorterStemmer) ResultBuffer() []rune { return p.b }

func (p *PorterStemmer) Stem(wordBuffer []rune) bool {
	p.reset()
	wordLen := len(wordBuffer)
	if len(p.b) < wordLen {
		p.b = make([]rune, util.Oversize(wordLen, util.NUM_BYTES_CHAR))
	}

	copy(p.b, wordBuffer)
	p.i = wordLen
	return p.stem(0)
}

func (p *PorterStemmer) cons(i int) bool {
	switch p.b[p.i] {
	case 'a':
	case 'e':
	case 'i':
	case 'o':
	case 'u':
		return false
	case 'y':
		if p.i == p.k0 {
			return true
		}

		return !p.cons(p.i - 1)
	default:
		return true
	}

	return true
}

func (p *PorterStemmer) m() int {
	n := 0
	i := p.k0
	for {
		if i > p.j {
			return n
		}

		if !p.cons(i) {
			break
		}
		i++
	}
	i++
	for {
		for {
			if i > p.j {
				return n
			}
			if p.cons(i) {
				break
			}
			i++
		}
		i++
		n++
		for {
			if i > p.j {
				return n
			}
			if !p.cons(i) {
				break
			}
			i++
		}
		i++
	}
}

func (p *PorterStemmer) vowelinstem() bool {
	for i := p.k0; i <= p.j; i++ {
		if !p.cons(i) {
			return true
		}
	}

	return false
}

func (p *PorterStemmer) doublec(j int) bool {
	if p.j < p.k0+1 {
		return false
	}
	if p.b[p.j] != p.b[p.j-1] {
		return false
	}
	return p.cons(p.j)
}

func (p *PorterStemmer) cvc(i int) bool {
	if p.i < p.k0+2 || !p.cons(p.i) || p.cons(p.i-1) || !p.cons(p.i-2) {
		return false
	} else {
		ch := p.b[p.i]
		if ch == 'w' || ch == 'x' || ch == 'y' {
			return false
		}
	}
	return true
}

func (p *PorterStemmer) ends(s string) bool {
	l := len(s)
	o := p.k - l + 1
	if o < p.k0 {
		return false
	}
	for i := 0; i < l; i++ {
		if p.b[o+i] != rune(s[i]) {
			return false
		}
	}
	p.j = p.k - l
	return true
}

func (p *PorterStemmer) setto(s string) {
	l := len(s)
	o := p.j + 1
	for i := 0; i < l; i++ {
		p.b[o+i] = rune(s[i])
	}
	p.k = p.j + l
	p.dirty = true
}

func (p *PorterStemmer) r(s string) {
	if p.m() > 0 {
		p.setto(s)
	}
}

func (p *PorterStemmer) step1() {
	if p.b[p.k] == 's' {
		if p.ends("sses") {
			p.k -= 2
		} else if p.ends("ies") {
			p.setto("i")
		} else if p.b[p.k-1] != 's' {
			p.k--
		}
	}
	if p.ends("eed") {
		if p.m() > 0 {
			p.k--
		}

	} else if (p.ends("ed") || p.ends("ing")) && p.vowelinstem() {
		p.k = p.j
		if p.ends("at") {
			p.setto("ate")
		} else if p.ends("bl") {
			p.setto("ble")
		} else if p.ends("iz") {
			p.setto("ize")
		} else if p.doublec(p.k) {
			ch := p.b[p.k]
			p.k--
			if ch == 'l' || ch == 's' || ch == 'z' {
				p.k++
			}
		} else if p.m() == 1 && p.cvc(p.k) {
			p.setto("e")
		}

	}
}

/* step2() turns terminal y to i when there is another vowel in the stem. */
func (p *PorterStemmer) step2() {
	if p.ends("y") && p.vowelinstem() {
		p.b[p.k] = 'i'
		p.dirty = true
	}
}

/* step3() maps double suffices to single ones. so -ization ( = -ize plus
   -ation) maps to -ize etc. note that the string before the suffix must give
   m() > 0. */

func (p *PorterStemmer) step3() {
	if p.k == p.k0 {
		return /* For Bug 1 */
	}
	switch p.b[p.k-1] {
	case 'a':
		if p.ends("ational") {
			p.r("ate")
			break
		}
		if p.ends("tional") {
			p.r("tion")
			break
		}
		break
	case 'c':
		if p.ends("enci") {
			p.r("ence")
			break
		}
		if p.ends("anci") {
			p.r("ance")
			break
		}
		break
	case 'e':
		if p.ends("izer") {
			p.r("ize")
			break
		}
		break
	case 'l':
		if p.ends("bli") {
			p.r("ble")
			break
		}
		if p.ends("alli") {
			p.r("al")
			break
		}
		if p.ends("entli") {
			p.r("ent")
			break
		}
		if p.ends("eli") {
			p.r("e")
			break
		}
		if p.ends("ousli") {
			p.r("ous")
			break
		}
		break
	case 'o':
		if p.ends("ization") {
			p.r("ize")
			break
		}
		if p.ends("ation") {
			p.r("ate")
			break
		}
		if p.ends("ator") {
			p.r("ate")
			break
		}
		break
	case 's':
		if p.ends("alism") {
			p.r("al")
			break
		}
		if p.ends("iveness") {
			p.r("ive")
			break
		}
		if p.ends("fulness") {
			p.r("ful")
			break
		}
		if p.ends("ousness") {
			p.r("ous")
			break
		}
		break
	case 't':
		if p.ends("aliti") {
			p.r("al")
			break
		}
		if p.ends("iviti") {
			p.r("ive")
			break
		}
		if p.ends("biliti") {
			p.r("ble")
			break
		}
		break
	case 'g':
		if p.ends("logi") {
			p.r("log")
			break
		}
	}
}

/* step4() deals with -ic-, -full, -ness etc. similar strategy to step3. */

func (p *PorterStemmer) step4() {
	switch p.b[p.k] {
	case 'e':
		if p.ends("icate") {
			p.r("ic")
			break
		}
		if p.ends("ative") {
			p.r("")
			break
		}
		if p.ends("alize") {
			p.r("al")
			break
		}
		break
	case 'i':
		if p.ends("iciti") {
			p.r("ic")
			break
		}
		break
	case 'l':
		if p.ends("ical") {
			p.r("ic")
			break
		}
		if p.ends("ful") {
			p.r("")
			break
		}
		break
	case 's':
		if p.ends("ness") {
			p.r("")
			break
		}
		break
	}
}

/* step5() takes off -ant, -ence etc., in context <c>vcvc<v>. */

func (p *PorterStemmer) step5() {
	if p.k == p.k0 {
		return /* for Bug 1 */
	}
	switch p.b[p.k-1] {
	case 'a':
		if p.ends("al") {
			break
		}
		return
	case 'c':
		if p.ends("ance") {
			break
		}
		if p.ends("ence") {
			break
		}
		return
	case 'e':
		if p.ends("er") {
			break
		}
		return
	case 'i':
		if p.ends("ic") {
			break
		}
		return
	case 'l':
		if p.ends("able") {
			break
		}
		if p.ends("ible") {
			break
		}
		return
	case 'n':
		if p.ends("ant") {
			break
		}
		if p.ends("ement") {
			break
		}
		if p.ends("ment") {
			break
		}
		/* element etc. not stripped before the m */
		if p.ends("ent") {
			break
		}
		return
	case 'o':
		if p.ends("ion") && p.j >= 0 && (p.b[p.j] == 's' || p.b[p.j] == 't') {
			break
		}
		/* j >= 0 fixes Bug 2 */
		if p.ends("ou") {
			break
		}
		return
		/* takes care of -ous */
	case 's':
		if p.ends("ism") {
			break
		}
		return
	case 't':
		if p.ends("ate") {
			break
		}
		if p.ends("iti") {
			break
		}
		return
	case 'u':
		if p.ends("ous") {
			break
		}
		return
	case 'v':
		if p.ends("ive") {
			break
		}
		return
	case 'z':
		if p.ends("ize") {
			break
		}
		return
	default:
		return
	}
	if p.m() > 1 {
		p.k = p.j
	}
}

/* step6() removes a final -e if m() > 1. */

func (p *PorterStemmer) step6() {
	p.j = p.k
	if p.b[p.k] == 'e' {
		a := p.m()
		if a > 1 || a == 1 && !p.cvc(p.k-1) {
			p.k--
		}

	}
	if p.b[p.k] == 'l' && p.doublec(p.k) && p.m() > 1 {
		p.k--
	}
}

func (p *PorterStemmer) stem(i0 int) bool {
	p.k = p.i - 1
	p.k0 = i0
	if p.k > p.k0+1 {
		p.step1()
		p.step2()
		p.step3()
		p.step4()
		p.step5()
		p.step6()
	}
	// Also, a word is considered dirty if we lopped off letters
	// Thanks to Ifigenia Vairelles for pointing this out.
	if p.i != p.k+1 {
		p.dirty = true
	}
	p.i = p.k + 1
	return p.dirty
}
