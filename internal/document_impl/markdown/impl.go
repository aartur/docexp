package markdown

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/aartur/docexp/internal/document"
	"github.com/aartur/docexp/internal/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const ExpansionStartTag = `<!-- <docexp> -->`
const ExpansionEndTag = `<!-- </docexp> -->`

var codeBlockMarker = []byte("```")

type SpanFinder struct {
	file     *utils.File
	rootNode ast.Node
}

func MakeSpanFinder(sourceFile *utils.File) (document.SpanFinder, error) {
	if err := sourceFile.EnsureFullContent(); err != nil {
		return &SpanFinder{}, err
	}
	md := goldmark.New(goldmark.WithParserOptions(parser.WithEscapedSpace()))
	reader := text.NewReader(sourceFile.FullContent)
	rootNode := md.Parser().Parse(reader)
	return &SpanFinder{
		file:     sourceFile,
		rootNode: rootNode,
	}, nil
}

func (f *SpanFinder) nodeMetadata(node ast.Node, headingsPath []*ast.Heading) utils.JSONObject {
	result := utils.JSONObject{}

	if f.file.Filename != nil {
		result["filename"] = f.file.Filename
	}

	filePos, err := f.file.FilePositionFromByteIndex(node.Pos())
	if err == nil {
		result["line"] = filePos.Line
		result["column"] = filePos.Column
	}

	if len(headingsPath) > 0 {
		headings := map[string]string{}
		for _, heading := range headingsPath {
			headings[strconv.Itoa(heading.Level)] = f.lineContent(heading)
		}
		result["headings"] = headings
	}

	return result
}

func (f *SpanFinder) lineContent(node ast.Node) string {
	var builder strings.Builder
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		builder.WriteString(string(line.Value(f.file.FullContent)))
	}
	return builder.String()
}

func (f *SpanFinder) nodeLen(node ast.Node) int {
	segments := node.Lines()
	if segments == nil {
		return 0
	}
	res := 0
	for i := 0; i < segments.Len(); i++ {
		seg := segments.At(i)
		res += seg.Len()
	}
	return res
}

func (f *SpanFinder) nodeContentEnd(node ast.Node) *int {
	segments := node.Lines()
	if segments == nil || segments.Len() == 0 {
		return nil
	}
	return new(segments.At(segments.Len() - 1).Stop)
}

func (f *SpanFinder) findCodeBlockStart(contentStartPos int) *int {
	idx := contentStartPos
	for {
		if idx < 3 {
			return nil
		}

		if slices.Equal(f.file.FullContent[idx-3:idx], codeBlockMarker) {
			return new(idx - 3)
		}

		// is it safe?
		if unicode.IsSpace(rune(f.file.FullContent[idx])) {
			idx--
		}

		return nil
	}
}

func (f *SpanFinder) findCodeBlockEnd(contentEndPos int) *int {
	idx := contentEndPos
	for {
		if idx > len(f.file.FullContent)-3 {
			return nil
		}

		nextTriple := f.file.FullContent[idx : idx+3]
		if slices.Equal(nextTriple, codeBlockMarker) {
			return new(idx + 3)
		}

		// is it safe?
		if unicode.IsSpace(rune(f.file.FullContent[idx])) {
			idx++
			continue
		}

		return nil
	}
}

type walkAlgorithm struct {
	spanFinder      *SpanFinder
	headingsPath    []*ast.Heading
	expStart        *int
	expContentStart *int
	walkError       error
	result          []document.Span
}

func (w *walkAlgorithm) Walk() {
	err := ast.Walk(w.spanFinder.rootNode, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		//fmt.Printf("Node Type: %s\n", n.Kind().String())

		switch node := n.(type) {
		case *ast.HTMLBlock:
			status, err := w.visitHTMLBlock(node)
			if err != nil {
				return status, err
			}
		case *ast.Heading:
			w.visitHeading(node)
		case *ast.FencedCodeBlock:
			// don't process code blocks which are inside expansion spans - docexp tags
			// cannot be nested
			if w.expContentStart != nil {
				return ast.WalkContinue, nil
			}
			status, err := w.visitFencedCodeBlock(node)
			if err != nil {
				return status, err
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		w.walkError = err
	}
}

func (w *walkAlgorithm) visitFencedCodeBlock(node *ast.FencedCodeBlock) (ast.WalkStatus, error) {
	metadata := w.spanFinder.nodeMetadata(node, w.headingsPath)
	content := w.spanFinder.lineContent(node)
	lang := node.Language(w.spanFinder.file.FullContent)
	if lang != nil {
		metadata["lang"] = string(lang)
	}

	// node.Pos() seems to return the index of the first backtick `
	// of the code block.
	startIndex := new(node.Pos())

	// it's harder to calculate the position of the closing backticks,
	// we need to manually scan the document content at the content end.
	contentEndIndex := w.spanFinder.nodeContentEnd(node)
	if contentEndIndex == nil {
		return ast.WalkStop, fmt.Errorf("cannot find code block content end after byte %d", node.Pos())
	}
	endIndex := w.spanFinder.findCodeBlockEnd(*contentEndIndex)
	if endIndex == nil {
		return ast.WalkStop, fmt.Errorf("cannot find code block end after byte %d", *contentEndIndex)
	}

	startPos, err := w.spanFinder.file.FilePositionFromByteIndex(*startIndex)
	if err != nil {
		return ast.WalkStop, err
	}
	endPos, err := w.spanFinder.file.FilePositionFromByteIndex(*endIndex - 1)
	if err != nil {
		return ast.WalkStop, err
	}
	span := document.Span{
		SourceSpan: document.SourceSpan{
			StartPos: startPos,
			EndPos:   endPos,
			Content:  content,
			Metadata: metadata,
		},
		ExpansionSpan: nil,
	}
	w.result = append(w.result, span)
	return 0, nil
}

func (w *walkAlgorithm) visitHeading(heading *ast.Heading) {
	// maintain headingsPath value
	lastValidIdx := len(w.headingsPath) - 1
	for {
		if lastValidIdx < 0 {
			break
		}
		if w.headingsPath[lastValidIdx].Level < heading.Level {
			break
		}
		lastValidIdx--
	}
	w.headingsPath = append(w.headingsPath[:lastValidIdx+1], heading)
}

func (w *walkAlgorithm) visitHTMLBlock(node *ast.HTMLBlock) (ast.WalkStatus, error) {
	content := w.spanFinder.lineContent(node)
	if utils.EqualCollapsingWhitespace(content, ExpansionStartTag) {
		if len(w.result) == 0 || w.result[len(w.result)-1].ExpansionSpan != nil {
			// detected expansion without a previous SourceSpan
			return ast.WalkStop, fmt.Errorf("invalid docexp tag opening: no expandable node above")
		}
		if w.expStart != nil {
			return ast.WalkStop, fmt.Errorf("docexp tags cannot be nested")
		}
		w.expStart = new(node.Pos())
		w.expContentStart = new(node.Pos() + len(content))
	} else if utils.EqualCollapsingWhitespace(content, ExpansionEndTag) {
		if w.expStart == nil {
			return ast.WalkStop, fmt.Errorf("unmatched closing docexp tag")
		}

		// detect if the HTML Block starts on a new line to strip unneeded newline
		// character
		var expContentEnd int
		if node.Pos() > 0 && w.spanFinder.file.FullContent[node.Pos()-1] == '\n' {
			expContentEnd = node.Pos() - 1
		} else {
			expContentEnd = node.Pos()
		}

		expContent := w.spanFinder.file.FullContent[*w.expContentStart:expContentEnd]
		startPos, err := w.spanFinder.file.FilePositionFromByteIndex(*w.expStart)
		if err != nil {
			return ast.WalkStop, fmt.Errorf("cannot get tag's start position")
		}
		endPos, err := w.spanFinder.file.FilePositionFromByteIndex(node.Pos() + len(content) - 1)
		if err != nil {
			return ast.WalkStop, fmt.Errorf("cannot get tag's end position")
		}
		expSpan := document.ExpansionSpan{StartPos: startPos, EndPos: endPos, Content: string(expContent)}

		lastResult := w.result[len(w.result)-1]
		lastResult.ExpansionSpan = &expSpan
		w.result[len(w.result)-1] = lastResult

		w.expStart = nil
		w.expContentStart = nil
	}
	return 0, nil
}

func (f *SpanFinder) FindSpans() ([]document.Span, error) {
	w := walkAlgorithm{spanFinder: f}
	w.Walk()
	if w.walkError != nil {
		return []document.Span{}, w.walkError
	}
	return w.result, nil
}

type ExpansionResultsSerializer struct{}

func (s *ExpansionResultsSerializer) Serialize(sourceFile *utils.File, results []document.ExpansionResult) (string, error) {
	if err := sourceFile.EnsureFullContent(); err != nil {
		return "", err
	}

	var builder strings.Builder
	copyFrom := 0
	for _, expRes := range results {
		if expRes.OriginalSpan.ExpansionSpan == nil && expRes.ExpansionContent == "" {
			continue
		}

		lang, err := s.getLang(expRes)
		if err != nil {
			return "", err
		}

		// copy the source span (we never modify the source spans)
		copyTo := expRes.OriginalSpan.SourceSpan.EndPos.Byte + 1
		precedingContent := sourceFile.FullContent[copyFrom:copyTo]
		builder.WriteString(string(precedingContent))

		// just write the new expanded content - when the source spans
		// are code blocks, we can just append the expansion
		builder.WriteString(s.formatExpansion(expRes.ExpansionContent, lang))
		// calculate copyFrom for the next loop step
		if expRes.OriginalSpan.ExpansionSpan == nil {
			copyFrom = expRes.OriginalSpan.SourceSpan.EndPos.Byte + 1
		} else {
			copyFrom = expRes.OriginalSpan.ExpansionSpan.EndPos.Byte + 1
		}

	}
	// write what's remaining
	if copyFrom < len(sourceFile.FullContent) {
		builder.WriteString(string(sourceFile.FullContent[copyFrom:]))
	}

	return builder.String(), nil
}

func (s *ExpansionResultsSerializer) getLang(expRes document.ExpansionResult) (string, error) {
	langAny, ok := expRes.Metadata["lang"]
	if !ok {
		return "", nil
	}
	var lang string
	switch v := langAny.(type) {
	case string:
		lang = v
	default:
		return "", fmt.Errorf("'lang' must be a string")
	}
	return lang, nil
}

func (s *ExpansionResultsSerializer) formatExpansion(content string, lang string) string {
	return "\n\n" +
		ExpansionStartTag + "\n" +
		s.formatCodeBlock(content, lang) +
		ExpansionEndTag + "\n"
}

func (s *ExpansionResultsSerializer) formatCodeBlock(content string, lang string) string {
	if content == "" || content[len(content)-1] != '\n' {
		content += "\n"
	}
	return "```" + lang + "\n" + content + "```" + "\n"
}
