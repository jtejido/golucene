package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_MU_DIRICHLET float32 = 2000
)

/**
 * Bayesian smoothing using Dirichlet priors. From Chengxiang Zhai and John
 * Lafferty. 2001. A study of smoothing methods for language models applied to
 * Ad Hoc information retrieval. In Proceedings of the 24th annual international
 * ACM SIGIR conference on Research and development in information retrieval
 * (SIGIR '01). ACM, New York, NY, USA, 334-342.
 * <p>
 * The formula as defined the paper assigns a negative score to documents that
 * contain the term, but with fewer occurrences than predicted by the collection
 * language model. The Lucene implementation returns {@code 0} for such
 * documents.
 * </p>
 *
 * @lucene.experimental
 */
type LMDirichletSimilarity struct {
	*lmSimilarityImpl
	mu float32
}

func NewLMDirichletSimilarity(mu float32) *LMDirichletSimilarity {
	ans := new(LMDirichletSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.mu = mu
	return ans
}

func NewLMDirichletSimilarityWithModel(collectionModel CollectionModel, mu float32) *LMDirichletSimilarity {
	ans := new(LMDirichletSimilarity)
	ans.lmSimilarityImpl = newLMSimilarity(ans, collectionModel)
	ans.mu = mu
	return ans
}

func NewDefaultLMDirichletSimilarity() *LMDirichletSimilarity {
	ans := new(LMDirichletSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.mu = DEFAULT_MU_DIRICHLET
	return ans
}

func (d *LMDirichletSimilarity) score(stats Stats, freq, docLen float32) float32 {
	score := stats.TotalBoost() * float32(math.Log(1+float64(freq/(d.mu*stats.(LMStats).CollectionProbability())))+math.Log(float64(d.mu/(docLen+d.mu))))

	if score > 0.0 {
		return score
	}

	return 0
}

func (d *LMDirichletSimilarity) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	expl.AddDetail(search.NewExplanation(d.mu, "mu"))
	weightExpl := search.NewExplanation(float32(math.Log(1+float64(freq/(d.mu*stats.(LMStats).CollectionProbability())))), "term weight")
	expl.AddDetail(weightExpl)
	expl.AddDetail(search.NewExplanation(float32(math.Log(float64(d.mu/(docLen+d.mu)))), "document norm"))
	d.lmSimilarityImpl.explain(expl, stats, doc, freq, docLen)
}

func (d *LMDirichletSimilarity) Name() string {
	return fmt.Sprintf("Dirichlet(%.2f)", d.mu)
}
