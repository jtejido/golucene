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
	"github.com/jtejido/golucene/queryparser/classic"
	"os"
)

func main() {
	util.SetDefaultInfoStream(util.NewPrintStreamInfoStream(os.Stdout))

	index.DefaultSimilarity = func() index.Similarity {
		return similarities.NewDefaultBM25Similarity()
	}

	directory, _ := store.OpenFSDirectory("test_index")
	analyzer := std.NewEnglishAnalyzer()
	conf := index.NewIndexWriterConfig(util.VERSION_LATEST, analyzer)
	writer, _ := index.NewIndexWriter(directory, conf)

	d := document.NewDocument()
	d.Add(document.NewTextFieldFromString("body", "test. 1 Lorem ipsum dolor sit amet, consectetur adipiscing elit.Sed vitae ante quis sem iaculis hendrerit.Interdum et malesuada fames ac ante ipsum primis in faucibus.Donec at luctus leo.Aenean eget tempor sem.Aliquam fermentum eleifend pretium.Sed fringilla, velit id mattis mattis, nisi elit consectetur sapien, id suscipit massa sem vitae justo", document.STORE_YES))
	writer.AddDocument(d.Fields())

	d2 := document.NewDocument()
	d2.Add(document.NewTextFieldFromString("body", "test 2 Lorem ipsum dolor sit amet, consectetur adipiscing elit.Sed vitae ante quis sem iaculis hendrerit.Interdum et malesuada fames ac ante ipsum primis in faucibus.Donec at luctus leo.Aenean eget tempor sem.Aliquam fermentum eleifend pretium.Sed fringilla, velit id mattis mattis, nisi elit consectetur sapien, id suscipit massa sem vitae justo", document.STORE_YES))
	writer.AddDocument(d2.Fields())

	writer.Close() // ensure index is written

	reader, _ := index.OpenDirectoryReader(directory)
	searcher := search.NewIndexSearcher(reader)

	parser := classic.NewQueryParser(util.VERSION_LATEST, "body", analyzer)

	var q search.Query
	var err error
	if q, err = parser.Parse(`test OR 1`); err != nil {
		fmt.Printf("error: %s", err.Error())
	}

	res, _ := searcher.Search(q, nil, 1000)
	fmt.Printf("Found %v hit(s).\n", res.TotalHits)
	for _, hit := range res.ScoreDocs {
		fmt.Printf("Doc %v score: %v\n", hit.Doc, hit.Score)
		doc, _ := reader.Document(hit.Doc)
		fmt.Printf("body -> %v\n", doc.Get("body"))
	}

	fmt.Printf("Searching via API - PhraseQuery.\n")
	q2 := search.NewPhraseQuery()
	q2.Add(index.NewTerm("body", "test"))
	q2.Add(index.NewTerm("body", "2"))

	res, _ = searcher.Search(q2, nil, 1000)
	fmt.Printf("Found %v hit(s).\n", res.TotalHits)
	for _, hit := range res.ScoreDocs {
		fmt.Printf("Doc %v score: %v\n", hit.Doc, hit.Score)
		doc, _ := reader.Document(hit.Doc)
		fmt.Printf("body -> %v\n", doc.Get("body"))
	}

}
