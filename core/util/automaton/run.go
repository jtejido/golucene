package automaton

import (
	"unicode"
)

// util/automaton/RunAutomaton.java

// Finite-state automaton with fast run operation.
type RunAutomaton struct {
	automaton   *Automaton
	maxInterval int
	size        int
	accept      []bool
	initial     int
	transitions []int // delta(state,c) = transitions[state*len(points)+CharClass(c)]
	points      []int // char interval start points
	classmap    []int // map from char number to class class
}

func (ra *RunAutomaton) String() string {
	panic("not implemented yet")
}

// Gets character class of given codepoint
func (ra *RunAutomaton) charClass(c int) int {
	return findIndex(c, ra.points)
}

func newRunAutomaton(a *Automaton, maxInterval int, tablesize bool) *RunAutomaton {
	return newRunAutomatonWithMaxDeterminedStates(a, maxInterval, tablesize, DEFAULT_MAX_DETERMINIZED_STATES)
}

// Constructs a new RunAutomaton from a deterministic Automaton.
func newRunAutomatonWithMaxDeterminedStates(a *Automaton, maxInterval int, tablesize bool, maxDeterminedStates int) *RunAutomaton {
	a = determinize(a, maxDeterminedStates)
	size := a.numStates()
	if size < 1 {
		size = 1
	}
	points := a.startPoints()
	nPoints := len(points)
	ans := &RunAutomaton{
		maxInterval: maxInterval,
		automaton:   a,
		points:      points,
		initial:     0,
		size:        size,
		accept:      make([]bool, size),
		transitions: make([]int, size*nPoints),
	}
	for i, _ := range ans.transitions {
		ans.transitions[i] = -1
	}
	for n := 0; n < size; n++ {
		ans.accept[n] = a.IsAccept(n)
		for c, point := range ans.points {
			dest := a.step(n, point)
			assert(dest == -1 || dest < size)
			ans.transitions[n*nPoints+c] = dest
		}
	}
	// Set alphabet table for optimal run performance.
	if tablesize {
		panic("not implemented yet")
	}
	return ans
}

/*
Returns the state obtained by reading the given char from the given
state. Returns -1 if not obtaining any such state. (If the original
Automaton had no dead states, -1 is returned here if and only if a
dead state is entered in an equivalent automaton with a total
transition function.)
*/
func (ra *RunAutomaton) step(state, c int) int {
	if ra.classmap == nil {
		return ra.transitions[state*len(ra.points)+ra.charClass(c)]
	} else {
		return ra.transitions[state*len(ra.points)+ra.classmap[c]]
	}
}

// Automaton representation for matching []char
type CharacterRunAutomaton struct {
	*RunAutomaton
}

func NewCharacterRunAutomaton(a *Automaton) *CharacterRunAutomaton {
	return NewCharacterRunAutomatonWithMaxDeterminedStates(a, DEFAULT_MAX_DETERMINIZED_STATES)
}

func NewCharacterRunAutomatonWithMaxDeterminedStates(a *Automaton, maxDeterminizedStates int) *CharacterRunAutomaton {
	ans := &CharacterRunAutomaton{}
	ans.RunAutomaton = newRunAutomatonWithMaxDeterminedStates(a, unicode.MaxRune, false, maxDeterminizedStates)
	return ans
}

func (ca *CharacterRunAutomaton) Run(s string) bool {
	p := ca.initial
	l := len(s)
	for i, cp := 0, 0; i < l; i++ {
		cp = int(s[i])
		p = ca.step(p, cp)

		if p == -1 {
			return false
		}
	}
	return ca.accept[p]
}

/**
 * Returns true if the given string is accepted by this automaton
 */
func (ca *CharacterRunAutomaton) RunChars(s []rune) {
	p := ca.initial
	l := len(s)
	for i, cp := 0, 0; i < l; i++ {
		cp = int(s[i])
		p = ca.step(p, cp)
		if p == -1 {
			return false
		}
	}
	return ca.accept[p]
}
