package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

/**
 * The Pitman-Yor Process (PYP) is used for probabilistic modeling of distributions that follow a power law.
 * Inference on a PYP can be efficiently approximated by combining power-law discounting with a Dirichlet-smoothed language model.
 *
 * From Saeedeh Momtazi, Dietrich Klakow. Hierarchical Pitman-Yor Language Model for Information Retrieval. SIGIR’10,July 19–23, 2010
 * From Antti Puurula. Cumulative Progress in Language Models for Information Retrieval. In Proceedings of Australasian Language Technology Association Workshop, pages 96−100.
 *
 * @lucene.experimental(jtejido)
 */
type LMPitmanYorProcessSimilarity struct {
	*lmSimilarityImpl
	mu    float32
	delta float32
}

func NewLMPitmanYorProcessSimilarity(mu, delta float32) *LMPitmanYorProcessSimilarity {
	ans := new(LMPitmanYorProcessSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.mu = mu
	ans.delta = delta
	return ans
}

func NewLMPitmanYorProcessSimilarityWithModel(collectionModel CollectionModel, mu, delta float32) *LMPitmanYorProcessSimilarity {
	ans := new(LMPitmanYorProcessSimilarity)
	ans.lmSimilarityImpl = newLMSimilarity(ans, collectionModel)
	ans.mu = mu
	ans.delta = delta
	return ans
}

func NewDefaultLMPitmanYorProcessSimilarity() *LMPitmanYorProcessSimilarity {
	ans := new(LMPitmanYorProcessSimilarity)
	ans.lmSimilarityImpl = newDefaultLMSimilarity(ans)
	ans.mu = DEFAULT_MU_DIRICHLET
	ans.delta = DEFAULT_DELTA_AD
	return ans
}

func (d *LMPitmanYorProcessSimilarity) score(stats Stats, freq, docLen float32) float32 {
	var tw float64
	if freq > 0 {
		tw = math.Pow(float64(freq), float64(d.delta))
	}

	freqPrime := float64(freq) - (float64(d.delta) * tw)

	if freqPrime < 0 {
		freqPrime = 0
	}

	score := stats.TotalBoost() * (float32(math.Log(1+(freqPrime/float64(d.mu*stats.(LMStats).CollectionProbability()))) + math.Log(1-float64(float32(stats.NumberOfFieldTokens())/(docLen+d.mu)))))

	if score > 0.0 {
		return score
	}

	return 0
}

func (d *LMPitmanYorProcessSimilarity) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	expl.AddDetail(search.NewExplanation(d.mu, "mu"))
	weightExpl := search.NewExplanation(float32(math.Log(1+float64(freq/(d.mu*stats.(LMStats).CollectionProbability())))), "term weight")
	expl.AddDetail(weightExpl)
	expl.AddDetail(search.NewExplanation(float32(math.Log(float64(d.mu/(docLen+d.mu)))), "document norm"))

}

func (d *LMPitmanYorProcessSimilarity) Name() string {
	return fmt.Sprintf("Pitman-Yor-Process(mu=%.2f, delta=%.2f)", d.mu, d.delta)
}
