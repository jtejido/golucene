package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

/**
 * XSqrAM is a class that computes the inner product of Pearson's X^2 with the information growth computed
 * with the multinomial M.
 *
 * It is an unsupervised DFR model of IR (free from parameters), which
 * can be used on short or medium verbose queries.
 *
 * Frequentist and Bayesian approach to  Information Retrieval. G. Amati. In Proceedings of the 28th European Conference on IR Research (ECIR 2006). LNCS vol 3936, pages 13--24.
 *
 */
type XSqrAMSimilarity struct {
	*LMSimilarity
}

func NewXSqrAMSimilarity(collectionModel CollectionModel) *XSqrAMSimilarity {
	ans := new(XSqrAMSimilarity)
	ans.LMSimilarity = newLMSimilarity(ans, collectionModel)

	return ans
}

func NewDefaultXSqrAMSimilarity() *XSqrAMSimilarity {
	ans := new(XSqrAMSimilarity)
	ans.LMSimilarity = newDefaultLMSimilarity(ans)

	return ans
}

func (xs *XSqrAMSimilarity) score(stats IBasicStats, freq, docLen float32) float32 {
	mle_d := float64(freq) / float64(docLen)
	smoothedProbability := (freq + 1) / (docLen + 1)
	mle_c := stats.CollectionProbability()
	XSqrA := float32(math.Pow(1-float64(mle_d), 2)) / (freq + 1)

	InformationDelta := (freq+1)*float32(math.Log(float64(smoothedProbability)/float64(mle_c))) - freq*float32(math.Log(float64(mle_d)/float64(mle_c))) + 0.5*float32(math.Log(float64(smoothedProbability)/float64(mle_d)))

	return freq * XSqrA * InformationDelta
}

func (xs *XSqrAMSimilarity) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}
}

func (xs *XSqrAMSimilarity) getName() string {
	return fmt.Sprintf("XSqrA_M")
}
