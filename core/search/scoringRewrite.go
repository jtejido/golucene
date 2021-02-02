package search

type ScoringRewrite interface {
	TermCollectingRewrite
	CheckMaxClauseCount(count int) error
}

type ScoringRewriteSPI interface {
	TopLevelQuery() Query
}

type AbstractScoringRewrite struct {
	spi ScoringRewriteSPI
}
