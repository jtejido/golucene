package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_LAMBDA_JM float32 = 0.7
)

/**
 * Language model based on the Jelinek-Mercer smoothing method. From Chengxiang
 * Zhai and John Lafferty. 2001. A study of smoothing methods for language
 * models applied to Ad Hoc information retrieval. In Proceedings of the 24th
 * annual international ACM SIGIR conference on Research and development in
 * information retrieval (SIGIR '01). ACM, New York, NY, USA, 334-342.
 * <p>The model has a single parameter, &lambda;. According to said paper, the
 * optimal value depends on both the collection and the query. The optimal value
 * is around {@code 0.1} for title queries and {@code 0.7} for long queries.</p>
 *
 * @lucene.experimental
 */
type LMJelinekMercerSimilarity struct {
	*lmSimilarityImpl
	lambda float32
}

func NewLMJelinekMercerSimilarityWithModel(collectionModel CollectionModel, lambda float32) *LMJelinekMercerSimilarity {
	ans := new(LMJelinekMercerSimilarity)
	ans.lmSimilarityImpl = newLMSimilarity(ans, collectionModel)
	ans.lambda = lambda
	return ans
}

func NewDefaultLMJelinekMercerSimilarity() *LMJelinekMercerSimilarity {
	ans := new(LMJelinekMercerSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.lambda = DEFAULT_LAMBDA_JM
	return ans
}

func NewLMJelinekMercerSimilarity(lambda float32) *LMJelinekMercerSimilarity {
	ans := new(LMJelinekMercerSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.lambda = lambda
	return ans
}

func (jm *LMJelinekMercerSimilarity) score(stats Stats, freq, docLen float32) float32 {
	return stats.TotalBoost() * float32(math.Log(1+((1-float64(jm.lambda))*float64(freq)/float64(docLen))/(float64(jm.lambda)*float64(stats.(LMStats).CollectionProbability())))+math.Log(float64(jm.lambda)))
}

func (jm *LMJelinekMercerSimilarity) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}
	expl.AddDetail(search.NewExplanation(jm.lambda, "lambda"))
	weightExpl := search.NewExplanation(float32(math.Log(1+((1-float64(jm.lambda))*float64(freq)/float64(docLen))/(float64(jm.lambda)*float64(stats.(LMStats).CollectionProbability())))), "term weight")
	expl.AddDetail(weightExpl)
	expl.AddDetail(search.NewExplanation(float32(math.Log(float64(jm.lambda))), "document norm"))
	jm.lmSimilarityImpl.explain(expl, stats, doc, freq, docLen)
}

func (jm *LMJelinekMercerSimilarity) Name() string {
	return fmt.Sprintf("Jelinek-Mercer(%.2f)", jm.lambda)
}
