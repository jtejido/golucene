package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_LAMBDA_HIEMSTRA float32 = 0.15
)

/**
 * HiemstraLM is a class for ranking documents against a query based on Hiemstra's PHD thesis for language
 * model.
 * @see https://pdfs.semanticscholar.org/67ba/b01706d3aada95e383f1296e5f019b869ae6.pdf
 *
 * @lucene.experimental(jtejido)
 */
type LMHiemstraSimilarity struct {
	*lmSimilarityImpl
	lambda float32
}

func NewLMHiemstraSimilarity(lambda float32) *LMHiemstraSimilarity {
	ans := new(LMHiemstraSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.lambda = lambda
	return ans
}

func NewLMHiemstraSimilarityWithModel(collectionModel CollectionModel, lambda float32) *LMHiemstraSimilarity {
	ans := new(LMHiemstraSimilarity)
	ans.lmSimilarityImpl = newLMSimilarity(ans, collectionModel)
	ans.lambda = lambda
	return ans
}

func NewDefaultLMHiemstraSimilarity() *LMHiemstraSimilarity {
	ans := new(LMHiemstraSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.lambda = DEFAULT_LAMBDA_HIEMSTRA
	return ans
}

func (d *LMHiemstraSimilarity) score(stats Stats, freq, docLen float32) float32 {
	score := stats.TotalBoost() * float32(math.Log(float64(1+((d.lambda*freq*float32(stats.NumberOfFieldTokens()))/((1-d.lambda)*float32(stats.TotalTermFreq())*docLen)))))

	if score > 0.0 {
		return score
	}

	return 0
}

func (d *LMHiemstraSimilarity) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	expl.AddDetail(search.NewExplanation(d.lambda, "lambda"))
	d.lmSimilarityImpl.explain(expl, stats, doc, freq, docLen)
}

func (d *LMHiemstraSimilarity) Name() string {
	return fmt.Sprintf("Hiemstra(%.2f)", d.lambda)
}
