package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_LAMBDA_JM float32 = 0.7
)

type LMJelinekMercerSimilarity struct {
	*LMSimilarity
	lambda float32
}

func NewLMJelinekMercerSimilarity(collectionModel CollectionModel, lambda float32) *LMJelinekMercerSimilarity {
	ans := new(LMJelinekMercerSimilarity)
	ans.LMSimilarity = newLMSimilarity(ans, collectionModel)
	ans.lambda = lambda
	return ans
}

func NewDefaultLMJelinekMercerSimilarity() *LMJelinekMercerSimilarity {
	ans := new(LMJelinekMercerSimilarity)
	ans.LMSimilarity = newDefaultLMSimilarity(ans)
	ans.lambda = DEFAULT_LAMBDA_JM
	return ans
}

func (jm *LMJelinekMercerSimilarity) score(stats IBasicStats, freq, docLen float32) float32 {
	return stats.TotalBoost() * float32(math.Log(1+((1-float64(jm.lambda))*float64(freq)/float64(docLen))/(float64(jm.lambda)*float64(stats.CollectionProbability()))))
}

func (jm *LMJelinekMercerSimilarity) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}
	expl.AddDetail(search.NewExplanation(jm.lambda, "lambda"))

}

func (jm *LMJelinekMercerSimilarity) getName() string {
	return fmt.Sprintf("Jelinek-Mercer(%.2f)", jm.lambda)
}
