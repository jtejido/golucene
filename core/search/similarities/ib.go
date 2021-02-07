package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
)

var _ Similarity = (*ibSimilarityImpl)(nil)
var _ IBSimilarity = (*ibSimilarityImpl)(nil)

type IBSimilarity interface {
	SimilarityBase
}

/**
 * Provides a framework for the family of information-based models, as described
 * in St&eacute;phane Clinchant and Eric Gaussier. 2010. Information-based
 * models for ad hoc IR. In Proceeding of the 33rd international ACM SIGIR
 * conference on Research and development in information retrieval (SIGIR '10).
 * ACM, New York, NY, USA, 234-241.
 * <p>The retrieval function is of the form <em>RSV(q, d) = &sum;
 * -x<sup>q</sup><sub>w</sub> log Prob(X<sub>w</sub> &ge;
 * t<sup>d</sup><sub>w</sub> | &lambda;<sub>w</sub>)</em>, where
 * <ul>
 *   <li><em>x<sup>q</sup><sub>w</sub></em> is the query boost;</li>
 *   <li><em>X<sub>w</sub></em> is a random variable that counts the occurrences
 *   of word <em>w</em>;</li>
 *   <li><em>t<sup>d</sup><sub>w</sub></em> is the normalized term frequency;</li>
 *   <li><em>&lambda;<sub>w</sub></em> is a parameter.</li>
 * </ul>
 * </p>
 * <p>The framework described in the paper has many similarities to the DFR
 * framework (see {@link DFRSimilarity}). It is possible that the two
 * Similarities will be merged at one point.</p>
 * <p>To construct an IBSimilarity, you must specify the implementations for
 * all three components of the Information-Based model.
 * <ol>
 *     <li>{@link Distribution}: Probabilistic distribution used to
 *         model term occurrence
 *         <ul>
 *             <li>{@link DistributionLL}: Log-logistic</li>
 *             <li>{@link DistributionLL}: Smoothed power-law</li>
 *         </ul>
 *     </li>
 *     <li>{@link Lambda}: &lambda;<sub>w</sub> parameter of the
 *         probability distribution
 *         <ul>
 *             <li>{@link LambdaDF}: <code>N<sub>w</sub>/N</code> or average
 *                 number of documents where w occurs</li>
 *             <li>{@link LambdaTTF}: <code>F<sub>w</sub>/N</code> or
 *                 average number of occurrences of w in the collection</li>
 *         </ul>
 *     </li>
 *     <li>{@link Normalization}: Term frequency normalization
 *         <blockquote>Any supported DFR normalization (listed in
 *                      {@link DFRSimilarity})</blockquote>
 *     </li>
 * </ol>
 * <p>
 * @see DFRSimilarity
 * @lucene.experimental
 */
type ibSimilarityImpl struct {
	*similarityBaseImpl
	distribution  Distribution
	lambda        Lambda
	normalization Normalization
}

func NewIBSimilarity(distribution Distribution, lambda Lambda, normalization Normalization) *ibSimilarityImpl {
	if distribution == nil || lambda == nil || normalization == nil {
		panic("nil  parameters not allowed.")
	}
	ans := &ibSimilarityImpl{distribution: distribution, lambda: lambda, normalization: normalization}
	ans.similarityBaseImpl = NewSimilarityBase(ans)
	return ans
}

func (ib *ibSimilarityImpl) score(stats Stats, freq, docLen float32) float32 {
	return stats.TotalBoost() * ib.distribution.Score(stats, ib.normalization.Tfn(stats, freq, docLen), ib.lambda.Lambda(stats))
}

func (ib *ibSimilarityImpl) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	if stats.TotalBoost() != 1.0 {
		expl.AddDetail(search.NewExplanation(stats.TotalBoost(), "boost"))
	}

	normExpl := ib.normalization.Explain(stats, freq, docLen)
	lambdaExpl := ib.lambda.Explain(stats)
	expl.AddDetail(normExpl)
	expl.AddDetail(lambdaExpl)
	expl.AddDetail(ib.distribution.Explain(stats, normExpl.Value(), lambdaExpl.Value()))
}

func (ib *ibSimilarityImpl) String() string {
	return fmt.Sprintf("IB %s-%s%s", ib.distribution.String(), ib.lambda.String(), ib.normalization.String())
}
