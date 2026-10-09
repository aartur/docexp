package markdown

import (
	"testing"

	"github.com/aartur/docexp/internal/document"
	"github.com/aartur/docexp/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestSpanFinder_noexp(t *testing.T) {
	sf, err := MakeSpanFinder(utils.NewFile("testdata/doc1_noexp.md"))
	if err != nil {
		t.Fatalf("SpanFinder creation: %v", err)
	}
	got, err := sf.FindSpans()
	if err != nil {
		t.Fatalf("Spans collection failed: %v", err)
	}
	//for _, span := range got {
	//	t.Logf("Got span\n%#v\n", span)
	//}
	want := []document.Span{
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 10, Column: 0, Byte: 53}, EndPos: utils.FilePosition{Line: 12, Column: 2, Byte: 76}, Content: "print(1+1)\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_noexp.md"), "headings": map[string]string{"1": "Header One", "2": "Header Two"}, "lang": "python", "line": 10}}, ExpansionSpan: (*document.ExpansionSpan)(nil)},
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 20, Column: 0, Byte: 118}, EndPos: utils.FilePosition{Line: 22, Column: 2, Byte: 133}, Content: "SELECT 1\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_noexp.md"), "headings": map[string]string{"1": "Header One", "2": "Header Four"}, "line": 20}}, ExpansionSpan: (*document.ExpansionSpan)(nil)},
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 28, Column: 0, Byte: 165}, EndPos: utils.FilePosition{Line: 30, Column: 2, Byte: 184}, Content: "uname -a\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_noexp.md"), "headings": map[string]string{"1": "Header Six"}, "lang": "bash", "line": 28}}, ExpansionSpan: (*document.ExpansionSpan)(nil)},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("wrong spans:\n%s", diff)
	}
}

func TestSpanFinder_exp(t *testing.T) {
	sf, err := MakeSpanFinder(utils.NewFile("testdata/doc1_exp.md"))
	if err != nil {
		t.Fatalf("SpanFinder creation: %v", err)
	}
	got, err := sf.FindSpans()
	if err != nil {
		t.Fatalf("Spans collection failed: %v", err)
	}
	//for _, span := range got {
	//	t.Logf("Got span\n%#v\n", span)
	//}
	want := []document.Span{
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 10, Column: 0, Byte: 53}, EndPos: utils.FilePosition{Line: 12, Column: 2, Byte: 76}, Content: "print(1+1)\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_exp.md"), "headings": map[string]string{"1": "Header One", "2": "Header Two"}, "lang": "python", "line": 10}}, ExpansionSpan: &document.ExpansionSpan{StartPos: utils.FilePosition{Line: 14, Column: 0, Byte: 79}, EndPos: utils.FilePosition{Line: 18, Column: 18, Byte: 125}, Content: "```\n2\n```"}},
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 26, Column: 0, Byte: 166}, EndPos: utils.FilePosition{Line: 28, Column: 2, Byte: 181}, Content: "SELECT 1\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_exp.md"), "headings": map[string]string{"1": "Header One", "2": "Header Four"}, "line": 26}}, ExpansionSpan: &document.ExpansionSpan{StartPos: utils.FilePosition{Line: 30, Column: 0, Byte: 184}, EndPos: utils.FilePosition{Line: 34, Column: 18, Byte: 230}, Content: "```\n1\n```"}},
		document.Span{SourceSpan: document.SourceSpan{StartPos: utils.FilePosition{Line: 40, Column: 0, Byte: 261}, EndPos: utils.FilePosition{Line: 42, Column: 2, Byte: 280}, Content: "uname -a\n", Metadata: utils.JSONObject{"column": 0, "filename": new("testdata/doc1_exp.md"), "headings": map[string]string{"1": "Header Six"}, "lang": "bash", "line": 40}}, ExpansionSpan: &document.ExpansionSpan{
			StartPos: utils.FilePosition{Line: 47, Column: 0, Byte: 315},
			EndPos:   utils.FilePosition{Line: 52, Column: 18, Byte: 420},
			Content:  "```\nLinux localhost 7.2 #1 SMP PREEMPT_DYNAMIC x86_64 GNU/Linux\n```\n",
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("wrong spans:\n%s", diff)
	}
}
