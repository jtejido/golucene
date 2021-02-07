package similarities

import (
	"github.com/jtejido/golucene/core/search"
	"math"
)

var (
	_ BasicModel = (*BasicModelBE)(nil)
	_ BasicModel = (*BasicModelD)(nil)
	_ BasicModel = (*BasicModelG)(nil)
	_ BasicModel = (*BasicModelIF)(nil)
	_ BasicModel = (*BasicModelIn)(nil)
	_ BasicModel = (*BasicModelIne)(nil)
	_ BasicModel = (*BasicModelP)(nil)
)

type BasicModel interface {
	Score(stats Stats, tfn float32) float32
	Explain(stats Stats, tfn float32) search.Explanation
	String() string
}

type basicModelSPI interface {
	Score(stats Stats, tfn float32) float32
}

type basicModel struct {
	owner basicModelSPI
}

func (bm *basicModel) Explain(stats Stats, tfn float32) search.Explanation {
	result := search.NewExplanation(bm.owner.Score(stats, tfn), ", computed from: ")
	result.AddDetail(search.NewExplanation(tfn, "tfn"))
	result.AddDetail(search.NewExplanation(float32(stats.NumberOfDocuments()), "numberOfDocuments"))
	result.AddDetail(search.NewExplanation(float32(stats.TotalTermFreq()), "totalTermFreq"))
	return result
}

type BasicModelBE struct {
	basicModel
}

func NewBasicModelBE() *BasicModelBE {
	ans := new(BasicModelBE)
	ans.owner = ans
	return ans
}

func (be *BasicModelBE) Score(stats Stats, tfn float32) float32 {
	F := float64(stats.TotalTermFreq()) + 1 + float64(tfn)
	// approximation only holds true when F << N, so we use N += F
	N := F + float64(stats.NumberOfDocuments())
	return float32(-math.Log2((N-1)*math.E) + be.f(N+F-1, N+F-float64(tfn)-2) - be.f(F, F-float64(tfn)))
}

func (be *BasicModelBE) f(n, m float64) float64 {
	return (m+0.5)*math.Log2(n/m) + (n-m)*math.Log2(n)
}

func (be *BasicModelBE) String() string {
	return "Be"
}

type BasicModelD struct {
	basicModel
}

func NewBasicModelD() *BasicModelD {
	ans := new(BasicModelD)
	ans.owner = ans
	return ans
}

func (bd *BasicModelD) Score(stats Stats, tfn float32) float32 {
	// we have to ensure phi is always < 1 for tiny TTF values, otherwise nphi can go negative,
	// resulting in NaN. cleanest way is to unconditionally always add tfn to totalTermFreq
	// to create a 'normalized' F.
	F := float64(stats.TotalTermFreq()) + 1 + float64(tfn)
	phi := float64(tfn) / F
	nphi := 1 - phi
	p := 1.0 / (float64(stats.NumberOfDocuments()) + 1)
	D := phi*math.Log2(phi/p) + nphi*math.Log2(nphi/(1-p))
	return float32(D*F + 0.5*math.Log2(1+2*math.Pi*float64(tfn)*nphi))
}

func (bd *BasicModelD) String() string {
	return "D"
}

type BasicModelG struct {
	basicModel
}

func NewBasicModelG() *BasicModelG {
	ans := new(BasicModelG)
	ans.owner = ans
	return ans
}

func (bg *BasicModelG) Score(stats Stats, tfn float32) float32 {
	// just like in BE, approximation only holds true when F << N, so we use lambda = F / (N + F)
	F := float64(stats.TotalTermFreq()) + 1
	N := float64(stats.NumberOfDocuments())
	lambda := F / (N + F)
	// -log(1 / (lambda + 1)) -> log(lambda + 1)
	return float32(math.Log2(lambda+1) + float64(tfn)*math.Log2((1+lambda)/lambda))
}

func (bg *BasicModelG) String() string {
	return "G"
}

type BasicModelIF struct {
	basicModel
}

func NewBasicModelIF() *BasicModelIF {
	ans := new(BasicModelIF)
	ans.owner = ans
	return ans
}

func (bg *BasicModelIF) Score(stats Stats, tfn float32) float32 {
	N := float64(stats.NumberOfDocuments())
	F := float64(stats.TotalTermFreq())
	return tfn * float32(math.Log2(1+(N+1)/(F+0.5)))
}

func (bg *BasicModelIF) String() string {
	return "I(F)"
}

type BasicModelIn struct {
	basicModel
}

func NewBasicModelIn() *BasicModelIn {
	ans := new(BasicModelIn)
	ans.owner = ans
	return ans
}

func (bg *BasicModelIn) Score(stats Stats, tfn float32) float32 {
	N := float64(stats.NumberOfDocuments())
	n := float64(stats.DocFreq())
	return tfn * float32(math.Log2((N+1)/(n+0.5)))
}

func (bg *BasicModelIn) Explain(stats Stats, tfn float32) search.Explanation {
	result := search.NewExplanation(bg.Score(stats, tfn), ", computed from: ")
	result.AddDetail(search.NewExplanation(tfn, "tfn"))
	result.AddDetail(search.NewExplanation(float32(stats.NumberOfDocuments()), "numberOfDocuments"))
	result.AddDetail(search.NewExplanation(float32(stats.DocFreq()), "docFreq"))
	return result
}

func (bg *BasicModelIn) String() string {
	return "I(n)"
}

type BasicModelIne struct {
	basicModel
}

func NewBasicModelIne() *BasicModelIne {
	ans := new(BasicModelIne)
	ans.owner = ans
	return ans
}

func (bine *BasicModelIne) Score(stats Stats, tfn float32) float32 {
	N := float64(stats.NumberOfDocuments())
	F := float64(stats.TotalTermFreq())
	ne := N * (1 - math.Pow((N-1)/N, F))
	return tfn * float32(math.Log2((N+1)/(ne+0.5)))
}

func (bine *BasicModelIne) String() string {
	return "I(ne)"
}

type BasicModelP struct {
	basicModel
}

func NewBasicModelP() *BasicModelP {
	ans := new(BasicModelP)
	ans.owner = ans
	return ans
}

func (bp *BasicModelP) Score(stats Stats, tfn float32) float32 {
	lambda := float64(stats.TotalTermFreq()+1) / float64(stats.NumberOfDocuments()+1)
	return float32(float64(tfn)*math.Log2(float64(tfn)/lambda) + (lambda+1/(12*float64(tfn))-float64(tfn))*math.Log2E + 0.5*math.Log2(2*math.Pi*float64(tfn)))
}

func (bp *BasicModelP) String() string {
	return "P"
}
