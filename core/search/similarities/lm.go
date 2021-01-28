package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
)

type ILMSimilarity interface {
	getName() string
	score(IBasicStats, float32, float32) float32
	explain(search.ExplanationSPI, IBasicStats, int, float32, float32)
}

type IBaseLMSimilarity interface {
	ILMSimilarity
	newStats(field string, queryBoost float32) IBasicStats
	fillBasicStats(stats IBasicStats, collectionStats search.CollectionStatistics, termStats search.TermStatistics)
	String() string
}

type CollectionModel interface {
	computeProbability(stats IBasicStats) float32
	getName() string
}

type LMSimilarity struct {
	*SimilarityBase
	spi             ILMSimilarity
	collectionModel CollectionModel
}

func newDefaultLMSimilarity(spi ILMSimilarity) *LMSimilarity {
	return newLMSimilarity(spi, new(DefaultCollectionModel))
}

func newLMSimilarity(spi ILMSimilarity, collectionModel CollectionModel) *LMSimilarity {
	ans := &LMSimilarity{spi: spi, collectionModel: collectionModel}
	ans.SimilarityBase = newSimilarityBase(ans)
	return ans
}

func (lms *LMSimilarity) newStats(field string, queryBoost float32) IBasicStats {
	return newLMStats(field, queryBoost)
}

func (lms *LMSimilarity) fillBasicStats(sstats IBasicStats, collectionStats search.CollectionStatistics, termStats search.TermStatistics) {
	stats := sstats.(*LMStats)
	stats.collectionProbability = lms.collectionModel.computeProbability(sstats)
}

func (lms *LMSimilarity) score(stats IBasicStats, freq, docLen float32) float32 {
	return lms.spi.score(stats, freq, docLen)
}

func (lms *LMSimilarity) explain(expl search.ExplanationSPI, stats IBasicStats, doc int, freq, docLen float32) {
	lms.spi.explain(expl, stats, doc, freq, docLen)
	expl.AddDetail(search.NewExplanation(lms.collectionModel.computeProbability(stats), "collection probability"))
}

func (lms *LMSimilarity) getName() string {
	return lms.spi.getName()
}

func (lms *LMSimilarity) String() string {
	coll := lms.collectionModel.getName()
	if coll != "" {
		return fmt.Sprintf("LM %s - %s", lms.getName(), coll)
	} else {
		return fmt.Sprintf("LM %s", lms.getName())
	}
}

type ILMStats interface {
	CollectionProbability() float32
}

type LMStats struct {
	*BasicStats
	collectionProbability float32
}

func newLMStats(field string, queryBoost float32) *LMStats {
	ans := new(LMStats)
	ans.BasicStats = newBasicStats(ans, field, queryBoost)
	return ans
}

func (lms *LMStats) CollectionProbability() float32 {
	return lms.collectionProbability
}

type DefaultCollectionModel struct{}

func (dcm *DefaultCollectionModel) computeProbability(stats IBasicStats) float32 {
	return float32((stats.TotalTermFreq() + 1)) / float32((stats.NumberOfFieldTokens() + 1))
}

func (dcm *DefaultCollectionModel) getName() string {
	return ""
}
