package similarities

import (
	"fmt"
	"github.com/jtejido/golucene/core/search"
)

type LMSimilarity interface {
	SimilarityBase
	Name() string
}

type lmSimilaritySPI interface {
	Name() string
	score(stats Stats, freq, docLen float32) float32
}

type lmSimilarityImpl struct {
	*similarityBaseImpl
	owner           lmSimilaritySPI
	collectionModel CollectionModel
}

func newDefaultLMSimilarity(owner lmSimilaritySPI) *lmSimilarityImpl {
	return newLMSimilarity(owner, new(DefaultCollectionModel))
}

func newLMSimilarity(owner lmSimilaritySPI, collectionModel CollectionModel) *lmSimilarityImpl {
	ans := &lmSimilarityImpl{owner: owner, collectionModel: collectionModel}
	ans.similarityBaseImpl = NewSimilarityBase(ans)
	return ans
}

func (lms *lmSimilarityImpl) newStats(field string, queryBoost float32) Stats {
	return newLMStats(field, queryBoost)
}

func (lms *lmSimilarityImpl) fillStats(stats Stats, collectionStats search.CollectionStatistics, termStats search.TermStatistics) {
	lms.similarityBaseImpl.fillStats(stats, collectionStats, termStats)
	lmStats := stats.(LMStats)
	lmStats.SetCollectionProbability(lms.collectionModel.ComputeProbability(stats))
}

func (lms *lmSimilarityImpl) score(stats Stats, freq, docLen float32) float32 {
	return lms.owner.score(stats, freq, docLen)
}

func (lms *lmSimilarityImpl) explain(expl search.ExplanationSPI, stats Stats, doc int, freq, docLen float32) {
	expl.AddDetail(search.NewExplanation(lms.collectionModel.ComputeProbability(stats), "collection probability"))
}

func (lms *lmSimilarityImpl) Name() string {
	return lms.owner.Name()
}

func (lms *lmSimilarityImpl) String() string {
	coll := lms.collectionModel.Name()
	if coll != "" {
		return fmt.Sprintf("LM %s - %s", lms.Name(), coll)
	} else {
		return fmt.Sprintf("LM %s", lms.Name())
	}
}

var _ Stats = (*lmStatsImpl)(nil)

type LMStats interface {
	Stats
	CollectionProbability() float32
	SetCollectionProbability(collectionProbability float32)
}

type lmStatsImpl struct {
	basicStatsImpl
	collectionProbability float32
}

func newLMStats(field string, queryBoost float32) *lmStatsImpl {
	ans := new(lmStatsImpl)
	ans.field = field
	ans.queryBoost = queryBoost
	ans.totalBoost = queryBoost
	return ans
}

func (lms *lmStatsImpl) CollectionProbability() float32 {
	return lms.collectionProbability
}

func (lms *lmStatsImpl) SetCollectionProbability(collectionProbability float32) {
	lms.collectionProbability = collectionProbability
}

type CollectionModel interface {
	ComputeProbability(stats Stats) float32
	Name() string
}

type DefaultCollectionModel struct{}

func (dcm *DefaultCollectionModel) ComputeProbability(stats Stats) float32 {
	return float32((stats.TotalTermFreq() + 1)) / float32((stats.NumberOfFieldTokens() + 1))
}

func (dcm *DefaultCollectionModel) Name() string {
	return ""
}
