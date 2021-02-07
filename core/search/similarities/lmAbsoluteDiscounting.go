package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_DELTA_AD float32 = 0.7
)

/**
 * Language model based on the Absolute-Discounting smoothing method. From Chengxiang
 * Zhai and John Lafferty. 2001. A study of smoothing methods for language
 * models applied to Ad Hoc information retrieval. In Proceedings of the 24th
 * annual international ACM SIGIR conference on Research and development in
 * information retrieval (SIGIR '01). ACM, New York, NY, USA, 334-342.
 * The model ranks documents against a query by lowering down the probability of seen words by
 * subtracting a constant from their counts.
 *
 * The effect of this is that the events with the lowest counts are discounted relatively more than those with higher counts.
 *
 * @lucene.experimental(jtejido)
 */
type LMAbsoluteDiscountingSimilarity struct {
	*lmSimilarityImpl
	delta float32
}

func NewLMAbsoluteDiscountingSimilarity(delta float32) *LMAbsoluteDiscountingSimilarity {
	ans := new(LMAbsoluteDiscountingSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.delta = delta
	return ans
}

func NewDefaultLMAbsoluteDiscountingSimilarity() *LMAbsoluteDiscountingSimilarity {
	ans := new(LMAbsoluteDiscountingSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.delta = DEFAULT_DELTA_AD
	return ans
}

func NewLMAbsoluteDiscountingSimilarityWithModel(collectionModel CollectionModel, delta float32) *LMAbsoluteDiscountingSimilarity {
	ans := new(LMAbsoluteDiscountingSimilarity)
	ans.lmSimilarityImpl = newLMSimilarity(ans, collectionModel)
	ans.delta = delta
	return ans
}

func (ad *LMAbsoluteDiscountingSimilarity) score(stats Stats, freq, docLen float32) float32 {
	return stats.TotalBoost() * float32(math.Log(1+float64((freq-ad.delta)/(ad.delta*float32(stats.NumberOfFieldTokens())*stats.(LMStats).CollectionProbability())))+math.Log(float64((ad.delta*float32(stats.NumberOfFieldTokens()))/docLen)))
}

func (ad *LMAbsoluteDiscountingSimilarity) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}
	expl.AddDetail(search.NewExplanation(ad.delta, "delta"))
	weightExpl := search.NewExplanation(float32(math.Log(1+float64((freq-ad.delta)/(ad.delta*float32(stats.NumberOfFieldTokens())*stats.(LMStats).CollectionProbability())))), "term weight")
	expl.AddDetail(weightExpl)
	expl.AddDetail(search.NewExplanation(float32(math.Log(float64((ad.delta*float32(stats.NumberOfFieldTokens()))/docLen))), "document norm"))
	ad.lmSimilarityImpl.explain(expl, stats, doc, freq, docLen)
}

func (ad *LMAbsoluteDiscountingSimilarity) Name() string {
	return fmt.Sprintf("AbsoluteDiscounting(%.2f)", ad.delta)
}
