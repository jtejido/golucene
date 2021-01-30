package search

type BoostedScorer struct {
  *FilterScorer
  boost float32
}

func newBoostedScorer(in Scorer, boost float32) (*BoostedScorer, error) {
  ans := &BoostedScorer{
    boost: boost,
  }
  ans.FilterScorer = newFilterScorer(ans)
  return ans, nil
}

func (s *BoostedScorer) Score() (sc float32, err error) {
  sc, err = s.in.Score()
  return sc * s.boost, err
}
