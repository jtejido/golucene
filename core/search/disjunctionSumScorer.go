package search

/** A Scorer for OR like queries, counterpart of <code>ConjunctionScorer</code>.
 * This Scorer implements {@link Scorer#advance(int)} and uses advance() on the given Scorers.
 */
type DisjunctionSumScorer struct {
	DisjunctionScorer
	score float64
	coord []float32
}

func newDisjunctionSumScorer(weight Weight, subScorers []Scorer, coord []float32) (*DisjunctionSumScorer, error) {
	ans := &DisjunctionSumScorer{
		DisjunctionScorer: DisjunctionScorer{
			doc:        -1,
			freq:       -1,
			subScorers: subScorers,
			numScorers: len(subScorers),
		},
		coord: coord,
	}

	ans.spi = ans
	ans.weight = weight

	ans.heapify()

	return ans, nil
}

func (s *DisjunctionSumScorer) reset() {
	s.score = 0
}

func (s *DisjunctionSumScorer) accum(subScorer Scorer) error {

	sc, err := subScorer.Score()
	if err != nil {
		return err
	}
	s.score += float64(sc)
	return nil
}

func (s *DisjunctionSumScorer) final() (float32, error) {
	return float32(s.score) * s.coord[s.freq], nil
}
