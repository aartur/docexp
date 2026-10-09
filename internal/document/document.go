package document

import (
	"github.com/aartur/docexp/internal/utils"
)

type SourceSpan struct {
	StartPos utils.FilePosition
	EndPos   utils.FilePosition
	Content  string

	Metadata utils.JSONObject
}

type ExpansionSpan struct {
	StartPos utils.FilePosition
	EndPos   utils.FilePosition
	Content  string
}

type Span struct {
	SourceSpan    SourceSpan
	ExpansionSpan *ExpansionSpan
}

type SpanFinder interface {
	FindSpans() ([]Span, error)
}

type ExpansionResult struct {
	OriginalSpan     Span
	ExpansionContent string
	Metadata         utils.JSONObject
}

type Expander interface {
	Expand(spans []Span) ([]ExpansionResult, error)
}

type ExpansionResultsSerializer interface {
	Serialize(sourceFile *utils.File, results []ExpansionResult) (string, error)
}
