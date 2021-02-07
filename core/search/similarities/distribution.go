package similarities

import (
	"github.com/jtejido/golucene/core/search"
	"math"
)

type Distribution interface {
	Score(stats Stats, tfn, lambda float32) float32
	Explain(stats Stats, tfn, lambda float32) search.Explanation
	String() string
}

type distributionSPI interface {
	Score(stats Stats, tfn, lambda float32) float32
}

type distribution struct {
	owner distributionSPI
}

func (d *distribution) Explain(stats Stats, tfn, lambda float32) search.Explanation {
	return search.NewExplanation(d.owner.Score(stats, tfn, lambda), "")
}

type DistributionLL struct {
	distribution
}

func NewDistributionLL() *DistributionLL {
	ans := new(DistributionLL)
	ans.owner = ans
	return ans
}

func (d *DistributionLL) Score(stats Stats, tfn, lambda float32) float32 {
	return float32(-math.Log(float64(lambda / (tfn + lambda))))
}

func (d *DistributionLL) String() string {
	return "LL"
}

type DistributionSPL struct {
	distribution
}

func NewDistributionSPL() *DistributionSPL {
	ans := new(DistributionSPL)
	ans.owner = ans
	return ans
}

func (d *DistributionSPL) Score(stats Stats, tfn, lambda float32) float32 {
	if lambda == 1 {
		lambda = 0.99
	}

	return float32(-math.Log((math.Pow(float64(lambda), float64(tfn/(tfn+1))) - float64(lambda)) / (1 - float64(lambda))))
}

func (d *DistributionSPL) String() string {
	return "SPL"
}
