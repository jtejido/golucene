package similarities

import (
	"github.com/jtejido/golucene/core/search"
)

type Lambda interface {
	Lambda(stats Stats) float32
	Explain(stats Stats) search.Explanation
	String() string
}

type LambdaDF struct{}

func NewLambdaDF() *LambdaDF {
	return new(LambdaDF)
}

func (l *LambdaDF) Lambda(stats Stats) float32 {
	return (float32(stats.DocFreq()) + 1) / (float32(stats.NumberOfDocuments()) + 1)
}

func (l *LambdaDF) Explain(stats Stats) search.Explanation {
	result := search.NewExplanation(l.Lambda(stats), ", computed from: ")
	result.AddDetail(search.NewExplanation(float32(stats.DocFreq()), "docFreq"))
	result.AddDetail(search.NewExplanation(float32(stats.NumberOfDocuments()), "numberOfDocuments"))
	return result
}

func (l *LambdaDF) String() string {
	return "D"
}

type LambdaTTF struct{}

func NewLambdaTTF() *LambdaTTF {
	return new(LambdaTTF)
}

func (l *LambdaTTF) Lambda(stats Stats) float32 {
	return (float32(stats.TotalTermFreq()) + 1) / (float32(stats.NumberOfDocuments()) + 1)
}

func (l *LambdaTTF) Explain(stats Stats) search.Explanation {
	result := search.NewExplanation(l.Lambda(stats), ", computed from: ")
	result.AddDetail(search.NewExplanation(float32(stats.TotalTermFreq()), "totalTermFreq"))
	result.AddDetail(search.NewExplanation(float32(stats.NumberOfDocuments()), "numberOfDocuments"))
	return result
}

func (l *LambdaTTF) String() string {
	return "TTF"
}
