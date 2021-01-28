package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

/**
 * TwoStageLM is a class for ranking documents that explicitly captures the different influences of the query and document
 * collection on the optimal settings of retrieval parameters.
 * It involves two steps. Estimate a document language for the model, and Compute the query likelihood using the estimated
 * language model. (DirichletLM and JelinkedMercerLM)
 *
 * From Chengxiang Zhai and John Lafferty. 2002. Two-Stage Language Models for Information Retrieval.
 * @see http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.7.3316&rep=rep1&type=pdf
 *
 * In a nutshell, this is a generalization of JelinkedMercerLM and DirichletLM.
 * The default values used here are the same constants found from the two classes.
 * Thus, making λ = 1 and μ same value as DirichletLM Class resolves the score towards DirichletLM, while as making μ larger
 * and λ same value as JelinekMercerLM Class resolves the score towards JelinekMercerLM.
 */
type LMTwoStageSimilarity struct {
	*LMSimilarity
	mu, lambda float32
}

func NewLMTwoStageSimilarity(collectionModel CollectionModel, lambda, mu float32) *LMTwoStageSimilarity {
	ans := new(LMTwoStageSimilarity)
	ans.LMSimilarity = newLMSimilarity(ans, collectionModel)
	ans.lambda = lambda
	ans.mu = mu
	return ans
}

func NewDefaultLMTwoStageSimilarity() *LMTwoStageSimilarity {
	ans := new(LMTwoStageSimilarity)
	ans.LMSimilarity = newDefaultLMSimilarity(ans)
	ans.lambda = DEFAULT_LAMBDA_JM
	ans.mu = DEFAULT_MU_DIRICHLET
	return ans
}

/**
 * Smoothed p(w|d) is ((1 - λ)(c(w|d) + (μp(w|C))) / (|d| + μ)) + λp(w|C));
 * Document dependent constant (norm) is (1-λ)|d| + μ / (|d| + μ)
 *
 * The term weight in a form of KL divergence is given by p(w|Q)log(p(w|d)/αp(w|C)) + log α where:
 * p(w|d) = the document model.
 * p(w|C) = the collection model.
 * p(w|Q) = the query model.
 * α = document dependent constant
 *
 * Thus it becomes log(1 + (λc(w|d) / ((1-λ)|d| + μ)p(w|C)) + log((1-λ)|d| + μ / (|d| + μ)).
 **/
func (ts *LMTwoStageSimilarity) score(stats IBasicStats, freq, docLen float32) float32 {
	norm := ((1-ts.lambda)*docLen + ts.mu) / (docLen + ts.mu)
	return stats.TotalBoost() * float32(math.Log(1+float64((ts.lambda*freq)/(((1-ts.lambda)*docLen+ts.mu)*stats.CollectionProbability())))+math.Log(float64(norm)))
}

func (ts *LMTwoStageSimilarity) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}
	expl.AddDetail(search.NewExplanation(ts.lambda, "lambda"))
	expl.AddDetail(search.NewExplanation(ts.mu, "mu"))

}

func (ts *LMTwoStageSimilarity) getName() string {
	return fmt.Sprintf("Two-Stage(lambda=%.2f, mu=$.2f)", ts.lambda, ts.mu)
}
