package main

import (
	"fmt"
	std "github.com/jtejido/golucene/analysis/en"
	_ "github.com/jtejido/golucene/core/codec/lucene410"
	"github.com/jtejido/golucene/core/document"
	"github.com/jtejido/golucene/core/index"
	"github.com/jtejido/golucene/core/search"
	"github.com/jtejido/golucene/core/search/similarities"
	"github.com/jtejido/golucene/core/store"
	"github.com/jtejido/golucene/core/util"
	"os"
)

func main() {
	util.SetDefaultInfoStream(util.NewPrintStreamInfoStream(os.Stdout))
	index.DefaultSimilarity = func() index.Similarity {
		return similarities.NewDefaultXSqrAMSimilarity()
	}

	directory, _ := store.OpenFSDirectory("test_index")
	analyzer := std.NewEnglishAnalyzer()
	conf := index.NewIndexWriterConfig(util.VERSION_LATEST, analyzer)
	writer, _ := index.NewIndexWriter(directory, conf)

	d := document.NewDocument()
	d.Add(document.NewTextFieldFromString("body", "According to current live statistics at the time of editing this letter, Russia has been the third country in the world to be affected by COVID-19 with both new cases and death rates rising.It remains in a position of advantage due to the later onset of the viral spread within the country since the worldwide disease outbreak", document.STORE_YES))
	writer.AddDocument(d.Fields())

	d2 := document.NewDocument()
	d2.Add(document.NewTextFieldFromString("body", "this is test 2", document.STORE_YES))
	writer.AddDocument(d2.Fields())

	d3 := document.NewDocument()
	d3.Add(document.NewTextFieldFromString("body", "this is test 3", document.STORE_YES))
	writer.AddDocument(d3.Fields())

	d4 := document.NewDocument()
	d4.Add(document.NewTextFieldFromString("body", "this is test 4", document.STORE_YES))
	writer.AddDocument(d4.Fields())

	writer.Close() // ensure index is written

	reader, _ := index.OpenDirectoryReader(directory)
	searcher := search.NewIndexSearcher(reader)

	q := search.NewTermQuery(index.NewTerm("body", "test"))
	res, _ := searcher.Search(q, nil, 1000)
	fmt.Printf("Found %v hit(s).\n", res.TotalHits)
	for _, hit := range res.ScoreDocs {
		fmt.Printf("Doc %v score: %v\n", hit.Doc, hit.Score)
		doc, _ := reader.Document(hit.Doc)
		fmt.Printf("body -> %v\n", doc.Get("body"))
	}

}
