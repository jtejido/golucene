package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
	"math"
)

const (
	DEFAULT_MU_DIRICHLET float32 = 2000
)

type LMDirichletSimilarity struct {
	*LMSimilarity
	mu float32
}

func NewLMDirichletSimilarity(collectionModel CollectionModel, mu float32) *LMDirichletSimilarity {
	ans := new(LMDirichletSimilarity)
	ans.LMSimilarity = newLMSimilarity(ans, collectionModel)
	ans.mu = mu
	return ans
}

func NewDefaultLMDirichletSimilarity() *LMDirichletSimilarity {
	ans := new(LMDirichletSimilarity)
	ans.LMSimilarity = newDefaultLMSimilarity(ans)
	ans.mu = DEFAULT_MU_DIRICHLET
	return ans
}

func (d *LMDirichletSimilarity) score(stats IBasicStats, freq, docLen float32) float32 {
	score := stats.TotalBoost() * float32(math.Log(1+float64(freq/(d.mu*stats.CollectionProbability())))+math.Log(float64(d.mu/(docLen+d.mu))))

	if score > 0.0 {
		return score
	}

	return 0
}

func (d *LMDirichletSimilarity) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	expl.AddDetail(search.NewExplanation(d.mu, "mu"))
	weightExpl := search.NewExplanation(float32(math.Log(1+float64(freq/(d.mu*stats.CollectionProbability())))), "term weight")
	expl.AddDetail(weightExpl)
	expl.AddDetail(search.NewExplanation(float32(math.Log(float64(d.mu/(docLen+d.mu)))), "document norm"))

}

func (d *LMDirichletSimilarity) getName() string {
	return fmt.Sprintf("Dirichlet(%.2f)", d.mu)
}
