package search

import (
	"fmt"
	. "github.com/jtejido/golucene/core/index/model"
	. "github.com/jtejido/golucene/core/search/model"
)

// search/TermScorer.java
/** Expert: A <code>Scorer</code> for documents matching a <code>Term</code>.
 */
type TermScorer struct {
	abstractScorer
	docsEnum  DocsEnum
	docScorer SimScorer
}

func newTermScorer(weight Weight, td DocsEnum, docScorer SimScorer) *TermScorer {
	ans := &TermScorer{docsEnum: td, docScorer: docScorer}
	ans.weight = weight
	return ans
}

func (ts *TermScorer) DocId() int {
	return ts.docsEnum.DocId()
}

func (ts *TermScorer) Freq() (int, error) {
	return ts.docsEnum.Freq()
}

/**
 * Advances to the next document matching the query. <br>
 *
 * @return the document matching the query or NO_MORE_DOCS if there are no more documents.
 */
func (ts *TermScorer) NextDoc() (d int, err error) {
	return ts.docsEnum.NextDoc()
}

func (ts *TermScorer) Score() (s float32, err error) {
	assert(ts.DocId() != NO_MORE_DOCS)
	freq, err := ts.docsEnum.Freq()
	if err != nil {
		return 0, err
	}
	return ts.docScorer.Score(ts.docsEnum.DocId(), float32(freq)), nil
}

/*
Advances to the first match beyond the current whose document number
is greater than or equal to a given target.
*/
func (ts *TermScorer) Advance(target int) (int, error) {
	return ts.docsEnum.Advance(target)
}

func (ts *TermScorer) String() string {
	return fmt.Sprintf("scorer(%v)", ts.weight)
}

func (ts *TermScorer) Cost() int64 {
	return ts.docsEnum.Cost()
}
