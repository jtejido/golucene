package similarities

import (
	"github.com/jtejido/golucene/core/search"
)

type AfterEffect interface {
	Score(stats Stats, tfn float32) float32
	Explain(stats Stats, tfn float32) search.Explanation
	String() string
}

type afterEffect struct{}

func (ae *afterEffect) Score(stats Stats, tfn float32) float32 {
	return 1
}

func (ae *afterEffect) Explain(stats Stats, tfn float32) search.Explanation {
	return search.NewExplanation(1, "no aftereffect")
}

type AfterEffectB struct {
	basicModel
}

func NewAfterEffectB() *AfterEffectB {
	ans := new(AfterEffectB)
	return ans
}

func (ae *AfterEffectB) Score(stats Stats, tfn float32) float32 {
	F := float64(stats.TotalTermFreq()) + 1
	n := float64(stats.DocFreq()) + 1
	return float32((F + 1) / (n * (float64(tfn) + 1)))
}

func (ae *AfterEffectB) Explain(stats Stats, tfn float32) search.Explanation {
	result := search.NewExplanation(ae.Score(stats, tfn), ", computed from: ")
	result.AddDetail(search.NewExplanation(tfn, "tfn"))
	result.AddDetail(search.NewExplanation(float32(stats.TotalTermFreq()), "totalTermFreq"))
	result.AddDetail(search.NewExplanation(float32(stats.DocFreq()), "docFreq"))
	return result
}

func (ae *AfterEffectB) String() string {
	return "B"
}

type AfterEffectL struct {
	basicModel
}

func NewAfterEffectL() *AfterEffectL {
	ans := new(AfterEffectL)
	return ans
}

func (ae *AfterEffectL) Score(stats Stats, tfn float32) float32 {
	return 1 / (tfn + 1)
}

func (ae *AfterEffectL) Explain(stats Stats, tfn float32) search.Explanation {
	result := search.NewExplanation(ae.Score(stats, tfn), ", computed from: ")
	result.AddDetail(search.NewExplanation(tfn, "tfn"))
	return result
}

func (ae *AfterEffectL) String() string {
	return "L"
}
