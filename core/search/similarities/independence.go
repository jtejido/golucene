package similarities

import (
	"math"
)

type Independence interface {
	Score(freq, expected float32) float32
	String() string
}

type IndependenceChiSquared struct{}

func NewIndependenceChiSquared() *IndependenceChiSquared {
	return new(IndependenceChiSquared)
}

func (i *IndependenceChiSquared) Score(freq, expected float32) float32 {
	return (freq - expected) * (freq - expected) / expected
}

func (i *IndependenceChiSquared) String() string {
	return "ChiSquared"
}

type IndependenceSaturated struct{}

func NewIndependenceSaturated() *IndependenceSaturated {
	return new(IndependenceSaturated)
}

func (i *IndependenceSaturated) Score(freq, expected float32) float32 {
	return (freq - expected) / expected
}

func (i *IndependenceSaturated) String() string {
	return "Saturated"
}

type IndependenceStandardized struct{}

func NewIndependenceStandardized() *IndependenceStandardized {
	return new(IndependenceStandardized)
}

func (i *IndependenceStandardized) Score(freq, expected float32) float32 {
	return (freq - expected) / float32(math.Sqrt(float64(expected)))
}

func (i *IndependenceStandardized) String() string {
	return "Standardized"
}
