package api

import (
	"fmt"
	"os"

	"github.com/aartur/docexp/internal/document_impl/markdown"
	"github.com/aartur/docexp/internal/expander_impl"
)

func CalcExpanderScriptInput(sourceFilePath string) ([]string, error) {
	sourceFile, err := OpenDocumentFile(sourceFilePath)
	if err != nil {
		return nil, err
	}
	spanFinder, err := markdown.MakeSpanFinder(sourceFile)
	if err != nil {
		return nil, err
	}
	spans, err := spanFinder.FindSpans()
	if err != nil {
		return nil, err
	}
	expander := expander_impl.ExtScriptExpander{}
	lines, err := expander.CalcScriptInput(spans)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func ExpandIntoFile(
	sourceFilePath string,
	destFilePath string,
	expanderScriptName string,
	expanderScriptArgs []string,
	expanderTimeoutSeconds int) error {

	sourceFile, err := OpenDocumentFile(sourceFilePath)
	if err != nil {
		return err
	}

	sourceFileInfo, err := os.Stat(sourceFilePath)
	if err != nil {
		return err
	}

	spanFinder, err := markdown.MakeSpanFinder(sourceFile)
	if err != nil {
		return err
	}
	spans, err := spanFinder.FindSpans()
	if err != nil {
		return err
	}
	expander := expander_impl.ExtScriptExpander{
		ScriptName:     expanderScriptName,
		ScriptArgs:     expanderScriptArgs,
		TimeoutSeconds: expanderTimeoutSeconds,
	}
	expResults, err := expander.Expand(spans)
	if err != nil {
		return err
	}
	serializer := markdown.ExpansionResultsSerializer{}
	serResult, err := serializer.Serialize(sourceFile, expResults)
	if err != nil {
		return err
	}
	err = os.WriteFile(destFilePath, []byte(serResult), sourceFileInfo.Mode())
	if err != nil {
		return fmt.Errorf("cannot write output document %s: %v", destFilePath, err)
	}
	return nil
}
