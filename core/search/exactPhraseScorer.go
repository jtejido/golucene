package search

import (
	"github.com/jtejido/golucene/core/index/model"
	. "github.com/jtejido/golucene/core/search/model"
)

const (
	CHUNK = 4096
)

type ChunkState struct {
	posEnum  model.DocsAndPositionsEnum
	offset   int32
	posUpto  int
	posLimit int
	pos      int
	lastPos  int
}

func newChunkState(posEnum model.DocsAndPositionsEnum, offset int32) *ChunkState {
	return &ChunkState{
		posEnum: posEnum,
		offset:  offset,
	}
}

type ExactPhraseScorer struct {
	*abstractScorer
	endMinus1, gen, docID, freq int
	counts, gens                []int
	cost                        int64
	chunkStates                 []*ChunkState
	lead                        model.DocsAndPositionsEnum
	docScorer                   SimScorer
}

func newExactPhraseScorer(weight Weight, postings []*PostingsAndFreq, docScorer SimScorer) (*ExactPhraseScorer, error) {
	ans := &ExactPhraseScorer{
		docScorer:   docScorer,
		chunkStates: make([]*ChunkState, len(postings)),
		endMinus1:   len(postings) - 1,
		docID:       -1,
		lead:        postings[0].postings,
		cost:        postings[0].postings.Cost(),
		gens:        make([]int, CHUNK),
		counts:      make([]int, CHUNK),
	}

	for i := 0; i < len(postings); i++ {
		ans.chunkStates[i] = newChunkState(postings[i].postings, -postings[i].position)
	}
	ans.abstractScorer = newScorer(ans, weight)
	return ans, nil
}

func (s *ExactPhraseScorer) doNext(doc int) (dd int, err error) {
	for {
		// TODO: don't dup this logic from conjunctionscorer :)
	advanceHead:
		for {
			for i := 1; i < len(s.chunkStates); i++ {
				de := s.chunkStates[i].posEnum
				if de.DocId() < doc {
					var d int
					d, err = de.Advance(doc)
					if err != nil {
						return 0, err
					}

					if d > doc {
						// DocsEnum beyond the current doc - break and advance lead to the new highest doc.
						doc = d
						break advanceHead
					}
				}
			}
			// all DocsEnums are on the same doc
			if doc == NO_MORE_DOCS {
				return doc, nil
			}

			var f int
			f, err = s.phraseFreq()
			if err != nil {
				return 0, err
			}
			if f > 0 {
				return doc, nil // success: matches phrase
			} else {
				doc, err = s.lead.NextDoc() // doesn't match phrase
				if err != nil {
					return 0, err
				}
			}
		}
		// advance head for next iteration
		doc, err = s.lead.Advance(doc)
		if err != nil {
			return 0, err
		}
	}

	return
}

func (s *ExactPhraseScorer) Score() (float32, error) {
	return s.docScorer.Score(s.docID, float32(s.freq)), nil
}

func (s *ExactPhraseScorer) Freq() (n int, err error) {
	return s.freq, nil
}

func (s *ExactPhraseScorer) DocId() int {
	return s.docID
}

func (s *ExactPhraseScorer) NextDoc() (doc int, err error) {
	doc, err = s.lead.NextDoc()
	if err != nil {
		return 0, err
	}
	s.docID, err = s.doNext(doc)
	return s.docID, err
}

func (s *ExactPhraseScorer) Advance(target int) (doc int, err error) {
	doc, err = s.lead.Advance(target)
	if err != nil {
		return 0, err
	}
	s.docID, err = s.doNext(doc)
	return s.docID, err
}

func (s *ExactPhraseScorer) Cost() int64 {
	return s.cost
}

func (s *ExactPhraseScorer) phraseFreq() (v int, err error) {
	s.freq = 0

	// init chunks
	for i := 0; i < len(s.chunkStates); i++ {
		cs := s.chunkStates[i]
		cs.posLimit, err = cs.posEnum.Freq()
		if err != nil {
			return 0, err
		}
		var pos int
		pos, err = cs.posEnum.NextPosition()
		if err != nil {
			return 0, err
		}
		cs.pos = int(cs.offset) + pos
		cs.posUpto = 1
		cs.lastPos = -1
	}

	chunkStart := 0
	chunkEnd := CHUNK

	// process chunk by chunk
	var end bool

	// TODO: we could fold in chunkStart into offset and
	// save one subtract per pos incr

	for !end {

		s.gen++

		if s.gen == 0 {
			// wraparound
			s.gens = make([]int, len(s.gens))
			s.gen++
		}

		// first term
		{
			cs := s.chunkStates[0]
			for cs.pos < chunkEnd {
				if cs.pos > cs.lastPos {
					cs.lastPos = cs.pos
					posIndex := cs.pos - chunkStart
					s.counts[posIndex] = 1
					assert(s.gens[posIndex] != s.gen)
					s.gens[posIndex] = s.gen
				}

				if cs.posUpto == cs.posLimit {
					end = true
					break
				}
				cs.posUpto++
				var pos int
				pos, err = cs.posEnum.NextPosition()
				if err != nil {
					return 0, err
				}
				cs.pos = int(cs.offset) + pos
			}
		}

		// middle terms
		any := true
		for t := 1; t < s.endMinus1; t++ {
			cs := s.chunkStates[t]
			any = false
			for cs.pos < chunkEnd {
				if cs.pos > cs.lastPos {
					cs.lastPos = cs.pos
					posIndex := cs.pos - chunkStart
					if posIndex >= 0 && s.gens[posIndex] == s.gen && s.counts[posIndex] == t {
						// viable
						s.counts[posIndex]++
						any = true
					}
				}

				if cs.posUpto == cs.posLimit {
					end = true
					break
				}
				cs.posUpto++
				var pos int
				pos, err = cs.posEnum.NextPosition()
				if err != nil {
					return 0, err
				}
				cs.pos = int(cs.offset) + pos
			}

			if !any {
				break
			}
		}

		if !any {
			// petered out for this chunk
			chunkStart += CHUNK
			chunkEnd += CHUNK
			continue
		}

		// last term

		{
			cs := s.chunkStates[s.endMinus1]
			for cs.pos < chunkEnd {
				if cs.pos > cs.lastPos {
					cs.lastPos = cs.pos
					posIndex := cs.pos - chunkStart
					if posIndex >= 0 && s.gens[posIndex] == s.gen && s.counts[posIndex] == s.endMinus1 {
						s.freq++
					}
				}

				if cs.posUpto == cs.posLimit {
					end = true
					break
				}
				cs.posUpto++
				var pos int
				pos, err = cs.posEnum.NextPosition()
				if err != nil {
					return 0, err
				}
				cs.pos = int(cs.offset) + pos
			}
		}

		chunkStart += CHUNK
		chunkEnd += CHUNK
	}

	return s.freq, nil
}
