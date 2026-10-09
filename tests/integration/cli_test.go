package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "docexp")
	cmd := exec.Command("go", "build", "-o", binPath, "../../main.go")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build docexp binary: %v\n%s", err, string(out))
	}
	return binPath
}

func run(t *testing.T, expScript string, doc string, expectedDoc string) {
	docPath, out, err := executeTool(t, expScript, doc)
	if err != nil {
		t.Fatalf("tool call failed: %v:\n%s", err, out)
	}
	docBytes, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("cannot read test file: %v", err)
	}
	if diff := cmp.Diff(expectedDoc, string(docBytes)); diff != "" {
		t.Errorf("invalid document after run:\n%s\n\n\nINPUT:\n%s\n\n\nOUTPUT:\n%s\n\n\nEXPECTED:\n%s", diff, doc, string(docBytes), expectedDoc)
	}
}

func executeTool(t *testing.T, expScript string, doc string) (string, []byte, error) {
	bin := buildBinary(t)
	docPath := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(docPath, []byte(doc), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}
	cmd := exec.Command(bin, "run", "--expander", filepath.Join("./testdata", expScript), docPath)
	//fmt.Printf("Running %v\n", cmd)
	out, err := cmd.CombinedOutput()
	return docPath, out, err
}

func postprocess(doc string) string {
	return strings.ReplaceAll(doc, "TICKS", "```")
}

func TestRun_empty(t *testing.T) {
	run(t, "exp_cat.py", "", "")
}

func TestRun_oneBlockOnly(t *testing.T) {
	doc := postprocess(`TICKS
echo 1
TICKS`)
	exp := postprocess(`TICKS
echo 1
TICKS

<!-- <docexp> -->
TICKS
exp: echo 1
TICKS
<!-- </docexp> -->
`)
	run(t, "exp_prefix.py", doc, exp)
}

func TestRun_oneBlockAlreadyExpanded(t *testing.T) {
	doc := postprocess(`TICKS
echo 1
TICKS

<!-- <docexp> -->
TICKS
exp: echo 1
TICKS
<!-- </docexp> -->
`)
	exp := doc
	run(t, "exp_prefix.py", doc, exp)
}

func TestRun_oneBlockSurrounded(t *testing.T) {
	doc := postprocess(`# Some Title

AAA

TICKS
echo 1
echo 2
TICKS

# Some section
`)
	exp := postprocess(`# Some Title

AAA

TICKS
echo 1
echo 2
TICKS

<!-- <docexp> -->
TICKS
exp: echo 1
echo 2
TICKS
<!-- </docexp> -->


# Some section
`)
	run(t, "exp_prefix.py", doc, exp)
}

func TestRun_replaceExpanded(t *testing.T) {
	doc := postprocess(`# Some Title

AAA

TICKS
echo 1
echo 2
TICKS

<!-- <docexp> -->
TICKS
echo 1
echo 2
TICKS
<!-- </docexp> -->


# Some section
`)
	exp := postprocess(`# Some Title

AAA

TICKS
echo 1
echo 2
TICKS

<!-- <docexp> -->
TICKS
exp: echo 1
echo 2
TICKS
<!-- </docexp> -->


# Some section
`)
	run(t, "exp_prefix.py", doc, exp)
}

func TestRun_errorInExpanderScript(t *testing.T) {
	_, out, err := executeTool(t, "exp_bad.py", `
TICKS
echo 1
TICKS
`)
	if err == nil {
		t.Fatalf("no error raised on invalid expander script")
	}
	fmt.Printf("Got error %v and output:\n%s\n", err, out)
}
