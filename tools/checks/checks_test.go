package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/*
	The checks are a safety net, and a safety net nobody tests is a decoration.
	What is covered here is the logic that can be wrong quietly: a path rule
	that matches too much, a link parser that skips a broken link, a cleanup
	that would step outside the repository.
*/

func TestUnderPathDoesNotMatchANeighbour(t *testing.T) {
	const api = "example.com/m/internal/api"

	if !underPath(api, api) {
		t.Fatal("a package must be under itself")
	}
	if !underPath(api+"/http/handlers", api) {
		t.Fatal("a nested package must be under its parent")
	}
	// Plain prefix matching would make internal/apiserver look like a package
	// inside internal/api, and forbid an import nobody wrote.
	if underPath("example.com/m/internal/apiserver", api) {
		t.Fatal("internal/apiserver must not count as internal/api")
	}
	if underPath("example.com/m/internal", api) {
		t.Fatal("a parent must not count as its child")
	}
}

func TestImportChainReportsTheShortestRoute(t *testing.T) {
	packages := map[string]*goPackage{
		"a": {ImportPath: "a", Imports: []string{"b", "c"}},
		"b": {ImportPath: "b", Imports: []string{"d"}},
		"c": {ImportPath: "c", Imports: []string{"net"}},
		"d": {ImportPath: "d", Imports: []string{"e"}},
		"e": {ImportPath: "e", Imports: []string{"net"}},
	}

	chain := importChain(packages, "a", "net")

	// Breadth-first, so the reader is told the most direct route rather than
	// whichever one the walk happened to reach first.
	want := []string{"a", "c", "net"}
	if strings.Join(chain, ",") != strings.Join(want, ",") {
		t.Fatalf("chain = %v, want %v", chain, want)
	}
}

func TestImportChainSurvivesACycle(t *testing.T) {
	// Go forbids import cycles, but `go list -e` will happily report one from
	// a broken tree, and a walk that trusted the graph would hang forever.
	packages := map[string]*goPackage{
		"a": {ImportPath: "a", Imports: []string{"b"}},
		"b": {ImportPath: "b", Imports: []string{"a"}},
	}

	chain := importChain(packages, "a", "net")

	if len(chain) != 2 || chain[0] != "a" {
		t.Fatalf("chain = %v, want a fallback naming the endpoints", chain)
	}
}

func TestReachesUsesWholePathSegments(t *testing.T) {
	if !reaches([]string{"fmt", "example.com/m/internal/api/http"}, "example.com/m/internal/api") {
		t.Fatal("a nested dependency must count as reaching its parent")
	}
	if reaches([]string{"example.com/m/internal/apiserver"}, "example.com/m/internal/api") {
		t.Fatal("a neighbour must not count as reaching")
	}
}

func TestLocalTargetsFindsWhatIsWorthChecking(t *testing.T) {
	body := strings.Join([]string{
		"[relative](docs/gateway.md)",
		"![an image](docs/files/diagram.png)",
		"[rooted](/README.md)",
		"[with anchor](docs/protocol.md#frames)",
		`[with title](docs/ci.md "Continuous integration")`,
		"[angle bracketed](<docs/a file.md>)",
		"[escaped](docs/a%20file.md)",
		"[external](https://example.com/x.md)",
		"[secure](http://example.com)",
		"[mail](mailto:someone@example.com)",
		"[anchor only](#section)",
		"[phone](tel:+70000000000)",
	}, "\n")

	got := localTargets(body)

	want := []string{
		"docs/gateway.md",
		"docs/files/diagram.png",
		"/README.md",
		"docs/protocol.md",
		"docs/ci.md",
		"docs/a file.md",
		"docs/a%20file.md",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("localTargets() = %v\nwant %v", got, want)
	}
}

func TestResolvesReadsALeadingSlashAsRepositoryRoot(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "#")
	write(t, filepath.Join(root, "docs", "gateway.md"), "#")

	// A rooted link is how these documents are written and how GitHub reads
	// them; taking it as the filesystem root would fail every one of them.
	if !resolves(root, "docs/ci.md", "/README.md") {
		t.Fatal("/README.md must resolve from the repository root")
	}
	if !resolves(root, "docs/ci.md", "gateway.md") {
		t.Fatal("a sibling must resolve relative to the document")
	}
	if !resolves(root, "docs/ci.md", "../README.md") {
		t.Fatal("a parent-relative link must resolve")
	}
	if resolves(root, "docs/ci.md", "missing.md") {
		t.Fatal("a link to nothing must not resolve")
	}
}

func TestResolvesUnescapesAPathWithSpaces(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "docs", "a file.md"), "#")

	if !resolves(root, "README.md", "docs/a%20file.md") {
		t.Fatal("a percent-escaped space must resolve to the file on disk")
	}
}

func TestArtefactsFindsOnlyWhatIsThere(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "bin", "ft12-api.exe"), "x")
	write(t, filepath.Join(root, "coverage.out"), "x")
	write(t, filepath.Join(root, "ft12-gateway.log"), "x")
	write(t, filepath.Join(root, "keep.md"), "x")
	// Root only: a recursive sweep for *.log would take this with it, and a
	// cleanup that eats test data is worse than no cleanup.
	write(t, filepath.Join(root, "internal", "testdata", "capture.log"), "x")

	got, err := artefacts(root)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, target := range got {
		names = append(names, rel(root, target))
	}
	want := "bin|coverage.out|ft12-gateway.log"
	if strings.Join(names, "|") != want {
		t.Fatalf("artefacts() = %v, want %s", names, want)
	}
}

func TestArtefactsListsEachPathOnce(t *testing.T) {
	root := t.TempDir()
	// coverage.out is both a named artefact and a *.out match.
	write(t, filepath.Join(root, "coverage.out"), "x")

	got, err := artefacts(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("artefacts() = %v, want one entry", got)
	}
}

func TestWithinRefusesToStepOutside(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "repo")

	if !within(root, filepath.Join(root, "bin")) {
		t.Fatal("a path inside the repository must be allowed")
	}
	if !within(root, root) {
		t.Fatal("the root itself must be allowed")
	}
	if within(root, filepath.Join(root, "..", "elsewhere")) {
		t.Fatal("a path outside the repository must be refused")
	}
	if within(root, filepath.Join(string(filepath.Separator), "etc")) {
		t.Fatal("an unrelated absolute path must be refused")
	}
}

func TestNormalizeNewlinesMakesTheComparisonPlatformBlind(t *testing.T) {
	// A Windows checkout holds CRLF and gofmt emits LF; comparing raw bytes
	// would fail every file on one platform and none on the other.
	if normalizeNewlines("a\r\nb\rc\n") != "a\nb\nc\n" {
		t.Fatal("CR and CRLF must both normalise to LF")
	}
}

func TestRepoRootFindsTheModuleFromASubdirectory(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	// The test runs in tools/checks, so a root found here proves the walk up.
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repoRoot() = %q, which has no go.mod", root)
	}
	if filepath.Base(root) == "checks" {
		t.Fatal("repoRoot() returned the package directory rather than the module root")
	}
}

func TestFailuresStaySilentUntilThereIsSomething(t *testing.T) {
	var f failures

	if f.err("headline") != nil {
		t.Fatal("an empty run must not produce an error")
	}

	f.addf("second")
	f.addf("first")
	err := f.err("headline")
	if err == nil {
		t.Fatal("a failure must produce an error")
	}
	// Sorted, so the same broken tree reports in the same order twice running
	// and a diff between two CI runs means something changed.
	if !strings.Contains(err.Error(), "first\n  second") {
		t.Fatalf("err = %q, want the items sorted", err)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
