package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

type Normalization interface {
	Tfn(stats Stats, tf, len float32) float32
	Explain(stats Stats, tf, len float32) search.Explanation
	String() string
}

type normalizationSPI interface {
	Tfn(stats Stats, tf, len float32) float32
}

type normalization struct {
	owner normalizationSPI
}

func (n *normalization) Explain(stats Stats, tf, len float32) search.Explanation {
	result := search.NewExplanation(n.owner.Tfn(stats, tf, len), ", computed from: ")
	result.AddDetail(search.NewExplanation(tf, "tf"))
	result.AddDetail(search.NewExplanation(float32(stats.AvgFieldLength()), "avgFieldLength"))
	result.AddDetail(search.NewExplanation(len, "len"))
	return result
}

type NoNormalization struct {
	normalization
}

func NewNoNormalization(c float32) *NoNormalization {
	ans := new(NoNormalization)
	ans.owner = ans
	return ans
}

func (n *NoNormalization) Tfn(stats Stats, tf, len float32) float32 {
	return tf
}

func (n *NoNormalization) Explain(stats Stats, tf, len float32) search.Explanation {
	return search.NewExplanation(1, "no normalization")
}

func (n *NoNormalization) String() string {
	return ""
}

type NormalizationH1 struct {
	normalization
	c float32
}

func NewNormalizationH1(c float32) *NormalizationH1 {
	ans := new(NormalizationH1)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationH1() *NormalizationH1 {
	return NewNormalizationH1(1)
}

func (n *NormalizationH1) Tfn(stats Stats, tf, len float32) float32 {
	return tf * n.c * float32(stats.AvgFieldLength()) / len
}

func (n *NormalizationH1) String() string {
	return "1"
}

type NormalizationH2 struct {
	normalization
	c float32
}

func NewNormalizationH2(c float32) *NormalizationH2 {
	ans := new(NormalizationH2)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationH2() *NormalizationH2 {
	return NewNormalizationH2(1)
}

func (n *NormalizationH2) Tfn(stats Stats, tf, len float32) float32 {
	return float32(float64(tf) * math.Log2(float64(1+(n.c*float32(stats.AvgFieldLength()))/len)))
}

func (n *NormalizationH2) String() string {
	return "2"
}

type NormalizationH2Exp struct {
	normalization
	c float32
}

func NewNormalizationH2Exp(c float32) *NormalizationH2Exp {
	ans := new(NormalizationH2Exp)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationH2Exp() *NormalizationH2Exp {
	return NewNormalizationH2Exp(1)
}

func (n *NormalizationH2Exp) Tfn(stats Stats, tf, len float32) float32 {
	return float32(float64(tf) * math.Log(float64(1+(n.c*float32(stats.AvgFieldLength()))/len)))
}

func (n *NormalizationH2Exp) String() string {
	return "2exp"
}

type NormalizationH3 struct {
	normalization
	mu float32
}

func NewNormalizationH3(mu float32) *NormalizationH3 {
	ans := new(NormalizationH3)
	ans.owner = ans
	ans.mu = mu
	return ans
}

func NewDefaultNormalizationH3() *NormalizationH3 {
	return NewNormalizationH3(1000)
}

func (n *NormalizationH3) Tfn(stats Stats, tf, len float32) float32 {
	return n.mu * (tf + n.mu*float32(stats.TotalTermFreq()+1)/float32(stats.NumberOfFieldTokens()+1)) / (len + n.mu)
}

func (n *NormalizationH3) String() string {
	return fmt.Sprintf("3(%.2f)", n.mu)
}

type NormalizationBM25 struct {
	normalization
	c float32
}

func NewNormalizationBM25(c float32) *NormalizationBM25 {
	ans := new(NormalizationBM25)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationBM25() *NormalizationBM25 {
	return NewNormalizationBM25(0.75)
}

func (n *NormalizationBM25) Tfn(stats Stats, tf, len float32) float32 {
	return tf / (1 - n.c + n.c*(len/float32(stats.AvgFieldLength())))
}

func (n *NormalizationBM25) String() string {
	return fmt.Sprintf("BM25(%.2f)", n.c)
}

type NormalizationF struct {
	normalization
	c float32
}

func NewNormalizationF(c float32) *NormalizationF {
	ans := new(NormalizationF)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationF() *NormalizationF {
	return NewNormalizationF(2500)
}

func (n *NormalizationF) Tfn(stats Stats, tf, len float32) float32 {
	return tf * (n.c * len / float32(stats.AvgFieldLength()))
}

func (n *NormalizationF) String() string {
	return "F"
}

/**
 * This class implements the tf normalisation based on Jelinek-Mercer smoothing for language modelling.
 */
type NormalizationJ struct {
	normalization
	c float32
}

func NewNormalizationJ(c float32) *NormalizationJ {
	ans := new(NormalizationJ)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationJ() *NormalizationJ {
	return NewNormalizationJ(.20)
}

func (n *NormalizationJ) Tfn(stats Stats, tf, len float32) float32 {
	mle_c := float32((stats.TotalTermFreq() + 1)) / float32((stats.NumberOfFieldTokens() + 1))
	mle_d := tf / len
	return ((1-n.c)*mle_d + (n.c * mle_c)) * len
}

func (n *NormalizationJ) String() string {
	return fmt.Sprintf("J(%.2f)", n.c)
}

/**
 * This class implements the tf normalisation based on Jelinek-Mercer smoothing
 * for language modelling where collection model is given by document frequency
 * instead of term frequency.
 */
type NormalizationJn struct {
	normalization
	c float32
}

func NewNormalizationJn(c float32) *NormalizationJn {
	ans := new(NormalizationJn)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationJn() *NormalizationJn {
	return NewNormalizationJn(.20)
}

func (n *NormalizationJn) Tfn(stats Stats, tf, len float32) float32 {
	mle_c := float32((stats.DocFreq() + 1)) / float32((stats.NumberOfFieldTokens() + 1))
	mle_d := tf / len
	return ((1-n.c)*mle_d + (n.c * mle_c)) * len
}

func (n *NormalizationJn) String() string {
	return fmt.Sprintf("Jn(%.2f)", n.c)
}

/**
 * This class implements Term Frequency Normalisation via
 * Pareto Distributions.
 * Lucene calls it NormalizationZ but where it was derived (Terrier), it's NormalizationP
 */
type NormalizationP struct {
	normalization
	c float32
}

func NewNormalizationP(c float32) *NormalizationP {
	ans := new(NormalizationP)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationP() *NormalizationP {
	return NewNormalizationP(0.30)
}

func (n *NormalizationP) Tfn(stats Stats, tf, len float32) float32 {
	return float32(float64(tf) * math.Pow(float64(stats.AvgFieldLength())/float64(len), float64(n.c)))
}

func (n *NormalizationP) String() string {
	return fmt.Sprintf("P(%.2f)", n.c)
}

/**
 * This class implements Dirichlet Priors normalization
 */
type NormalizationDP struct {
	normalization
	c float32
}

func NewNormalizationDP(c float32) *NormalizationDP {
	ans := new(NormalizationDP)
	ans.owner = ans
	ans.c = c
	return ans
}

func NewDefaultNormalizationDP() *NormalizationDP {
	return NewNormalizationDP(2500)
}

func (n *NormalizationDP) Tfn(stats Stats, tf, len float32) float32 {
	mle_c := float32((stats.TotalTermFreq() + 1)) / float32((stats.NumberOfFieldTokens() + 1))
	return n.c * (tf + n.c*mle_c) / (len + n.c)
}

func (n *NormalizationDP) String() string {
	return fmt.Sprintf("DP(%.2f)", n.c)
}
