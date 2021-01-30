package search

type FilterScorer struct {
	*abstractScorer
	in Scorer
}

func newFilterScorer(in Scorer) *FilterScorer {
	ans := &FilterScorer{
		in: in,
	}
	ans.abstractScorer = newScorer(ans, in.Weight())
	return ans
}

func (s *FilterScorer) Score() (float32, error) {
	return s.in.Score()
}

func (s *FilterScorer) Freq() (n int, err error) {
	return s.in.Freq()
}

func (s *FilterScorer) DocId() int {
	return s.in.DocId()
}

func (s *FilterScorer) NextDoc() (doc int, err error) {
	return s.in.NextDoc()
}

func (s *FilterScorer) Advance(target int) (doc int, err error) {
	return s.in.Advance(target)
}

func (s *FilterScorer) Cost() int64 {
	return s.in.Cost()
}
