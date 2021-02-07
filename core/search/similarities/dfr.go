package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
)

var _ Similarity = (*dfrSimilarityImpl)(nil)
var _ DFRSimilarity = (*dfrSimilarityImpl)(nil)

type DFRSimilarity interface {
	SimilarityBase
}

/**
 * Implements the <em>divergence from randomness (DFR)</em> framework
 * introduced in Gianni Amati and Cornelis Joost Van Rijsbergen. 2002.
 * Probabilistic models of information retrieval based on measuring the
 * divergence from randomness. ACM Trans. Inf. Syst. 20, 4 (October 2002),
 * 357-389.
 * <p>The DFR scoring formula is composed of three separate components: the
 * <em>basic model</em>, the <em>aftereffect</em> and an additional
 * <em>normalization</em> component, represented by the classes
 * {@code BasicModel}, {@code AfterEffect} and {@code Normalization},
 * respectively. The names of these classes were chosen to match the names of
 * their counterparts in the Terrier IR engine.</p>
 * <p>To construct a DFRSimilarity, you must specify the implementations for
 * all three components of DFR:
 * <ol>
 *    <li>{@link BasicModel}: Basic model of information content:
 *        <ul>
 *           <li>{@link BasicModelBE}: Limiting form of Bose-Einstein
 *           <li>{@link BasicModelG}: Geometric approximation of Bose-Einstein
 *           <li>{@link BasicModelP}: Poisson approximation of the Binomial
 *           <li>{@link BasicModelD}: Divergence approximation of the Binomial
 *           <li>{@link BasicModelIn}: Inverse document frequency
 *           <li>{@link BasicModelIne}: Inverse expected document
 *               frequency [mixture of Poisson and IDF]
 *           <li>{@link BasicModelIF}: Inverse term frequency
 *               [approximation of I(ne)]
 *        </ul>
 *    <li>{@link AfterEffect}: First normalization of information
 *        gain:
 *        <ul>
 *           <li>{@link AfterEffectL}: Laplace's law of succession
 *           <li>{@link AfterEffectB}: Ratio of two Bernoulli processes
 *           <li>{@link NoAfterEffect}: no first normalization
 *        </ul>
 *    <li>{@link Normalization}: Second (length) normalization:
 *        <ul>
 *           <li>{@link NormalizationH1}: Uniform distribution of term
 *               frequency
 *           <li>{@link NormalizationH2}: term frequency density inversely
 *               related to length
 *           <li>{@link NormalizationH3}: term frequency normalization
 *               provided by Dirichlet prior
 *           <li>{@link NormalizationZ}: term frequency normalization provided
 *                by a Zipfian relation
 *           <li>{@link NoNormalization}: no second normalization
 *        </ul>
 * </ol>
 * <p>Note that <em>qtf</em>, the multiplicity of term-occurrence in the query,
 * is not handled by this implementation.</p>
 * @see BasicModel
 * @see AfterEffect
 * @see Normalization
 * @lucene.experimental
 */
type dfrSimilarityImpl struct {
	*similarityBaseImpl
	basicModel    BasicModel
	afterEffect   AfterEffect
	normalization Normalization
}

func NewDFRSimilarity(basicModel BasicModel, afterEffect AfterEffect, normalization Normalization) *dfrSimilarityImpl {
	if basicModel == nil || afterEffect == nil || normalization == nil {
		panic("nil  parameters not allowed.")
	}
	ans := &dfrSimilarityImpl{basicModel: basicModel, afterEffect: afterEffect, normalization: normalization}
	ans.similarityBaseImpl = NewSimilarityBase(ans)
	return ans
}

func (dfr *dfrSimilarityImpl) score(stats Stats, freq, docLen float32) float32 {
	tfn := dfr.normalization.Tfn(stats, freq, docLen)
	return stats.TotalBoost() * dfr.basicModel.Score(stats, tfn) * dfr.afterEffect.Score(stats, tfn)
}

func (dfr *dfrSimilarityImpl) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	normExpl := dfr.normalization.Explain(stats, freq, docLen)
	tfn := normExpl.Value()
	expl.AddDetail(normExpl)
	expl.AddDetail(dfr.basicModel.Explain(stats, tfn))
	expl.AddDetail(dfr.afterEffect.Explain(stats, tfn))
}

func (dfr *dfrSimilarityImpl) String() string {
	return fmt.Sprintf("DFR %s %s %s", dfr.basicModel.String(), dfr.afterEffect.String(), dfr.normalization.String())
}
