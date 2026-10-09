package expander_impl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aartur/docexp/internal/document"
)

type ExtScriptExpander struct {
	ScriptName     string
	ScriptArgs     []string
	TimeoutSeconds int
}

type expanderRun struct {
	expander         ExtScriptExpander
	spans            []document.Span
	scriptInputLines []string
	results          []document.ExpansionResult
}

func (r *expanderRun) calcInputLines() error {
	r.scriptInputLines = nil
	for i, span := range r.spans {
		spanId := i + 1

		lineObj := map[string]any{}
		lineObj["id"] = spanId
		lineObj["content"] = span.SourceSpan.Content
		lineObj["metadata"] = span.SourceSpan.Metadata
		lineStr, err := json.Marshal(lineObj)
		if err != nil {
			return fmt.Errorf("cannot serialize metadata of span %d: %w", spanId, err)
		}

		r.scriptInputLines = append(r.scriptInputLines, string(lineStr))

	}
	return nil

	//if s.TimeoutPerCallSeconds > 0 {
	//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//	defer cancel()
	//}
}

func (r *expanderRun) parseOutput(output *string) error {
	clear(r.results)

	lines := strings.Split(*output, "\n")
	for i, line := range lines {
		lineNo := i + 1

		if strings.TrimSpace(line) == "" {
			continue
		}

		lineBytes := []byte(line)
		var lineObj map[string]any
		err := json.Unmarshal(lineBytes, &lineObj)
		if err != nil {
			return fmt.Errorf("cannot unmarshal JSON object from line %d of output: %w", lineNo, err)
		}

		idAny, ok := lineObj["id"]
		if !ok {
			return fmt.Errorf("JSON object from line %d doesn't contain the required key 'id'", lineNo)
		}
		if idAny == nil {
			return fmt.Errorf("id cannot be null at line %d", lineNo)
		}
		var id int
		switch v := idAny.(type) {
		case float64:
			id = int(v)
		default:
			return fmt.Errorf("'id' must be an integer at line %d", lineNo)
		}
		if id < 1 || id > len(r.results) {
			return fmt.Errorf("'id' is out of range at line %d", lineNo)
		}
		index := id - 1

		expansionAny, ok := lineObj["exp"]
		if !ok {
			return fmt.Errorf("no required key 'exp' at line %d", lineNo)
		}
		var expansion string
		if expansionAny == nil {
			expansion = ""
		} else {
			switch v := expansionAny.(type) {
			case string:
				expansion = v
			default:
				return fmt.Errorf("'exp' must be a string at line %d", lineNo)
			}
		}

		// empty content means we want to have the expansion cleared
		r.results[index] = document.ExpansionResult{
			OriginalSpan:     r.spans[index],
			ExpansionContent: expansion,
		}
	}
	return nil
}

func initExpanderRun(e ExtScriptExpander, spans []document.Span) (*expanderRun, error) {
	run := &expanderRun{
		expander: e,
		spans:    spans,
		results:  make([]document.ExpansionResult, len(spans)),
	}
	err := run.calcInputLines()
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (r *expanderRun) runExpanderScript() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.expander.TimeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.expander.ScriptName, r.expander.ScriptArgs...)

	scriptInput := strings.Join(r.scriptInputLines, "\n")
	cmd.Stdin = bytes.NewBufferString(scriptInput)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", errors.New("timeout when running the expansion script")
		}

		return "", fmt.Errorf("when calling the expansion command: %v; stderr:\n%s)", err, stderrBuf.String())
	}

	stdoutStr := stdoutBuf.String()
	return stdoutStr, nil
}

func (e ExtScriptExpander) Expand(spans []document.Span) ([]document.ExpansionResult, error) {
	run, err := initExpanderRun(e, spans)
	if err != nil {
		return nil, err
	}
	out, err := run.runExpanderScript()
	if err != nil {
		return []document.ExpansionResult{}, err
	}
	err = run.parseOutput(&out)
	if err != nil {
		return []document.ExpansionResult{}, err
	}
	return run.results, nil
}

func (e ExtScriptExpander) CalcScriptInput(spans []document.Span) ([]string, error) {
	run, err := initExpanderRun(e, spans)
	if err != nil {
		return nil, err
	}
	err = run.calcInputLines()
	if err != nil {
		return nil, err
	}
	return run.scriptInputLines, nil
}
