package main

import (
	"os"
	"os/exec"
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

func TestGeneratedRecognisesProtobufOutput(t *testing.T) {
	if !generated("internal/api/grpc/ft12/v1/gateway.pb.go") {
		t.Fatal("a .pb.go file is generated")
	}
	if !generated("internal/api/grpc/ft12/v1/gateway_grpc.pb.go") {
		t.Fatal("a _grpc.pb.go file is generated")
	}
	// Nothing else: excluding a hand-written file from coverage because its
	// name looks similar would hide exactly the code the number is about.
	if generated("internal/gateway/settings.go") {
		t.Fatal("a hand-written file must be counted")
	}
	if generated("internal/protocol/frame/pb.go") {
		t.Fatal("only the .pb.go suffix counts, not a file called pb.go")
	}
}

func TestMergeFoldsRepeatedSpans(t *testing.T) {
	// What -coverpkg=./... produces: every test binary reports on every
	// package, so one span arrives once per binary. Summing them counted a
	// module of 4,700 statements as one of 122,000 and reported six percent
	// coverage of a well-tested tree.
	merged := merge([]block{
		{file: "a.go", span: "1.1,2.2", statements: 3, count: 0},
		{file: "a.go", span: "1.1,2.2", statements: 3, count: 7},
		{file: "a.go", span: "1.1,2.2", statements: 3, count: 0},
		{file: "a.go", span: "9.1,9.9", statements: 1, count: 0},
	})

	if len(merged) != 2 {
		t.Fatalf("merge() kept %d blocks, want 2", len(merged))
	}
	// A statement any binary reached is reached.
	if merged[0].count != 7 {
		t.Fatalf("count = %d, want the highest of the reports", merged[0].count)
	}
	if merged[1].count != 0 {
		t.Fatalf("an uncovered span must stay uncovered, got %d", merged[1].count)
	}
}

func TestMergeKeepsSpansApartWithinAFile(t *testing.T) {
	merged := merge([]block{
		{file: "a.go", span: "1.1,2.2", statements: 3, count: 1},
		{file: "a.go", span: "3.1,4.2", statements: 5, count: 0},
		{file: "b.go", span: "1.1,2.2", statements: 2, count: 0},
	})

	if len(merged) != 3 {
		t.Fatalf("merge() kept %d blocks, want 3", len(merged))
	}
}

func TestCoveragePercentOfNothing(t *testing.T) {
	// A profile with no statements is not zero percent covered; it is a
	// question with no answer, and dividing by zero would say NaN.
	if got := (coverage{}).percent(); got != 0 {
		t.Fatalf("percent() = %v", got)
	}
}

func TestReadProfileParsesAndSurvivesRubbish(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "cover.out")
	write(t, path, strings.Join([]string{
		"mode: atomic",
		"example.com/m/a.go:10.20,12.3 4 1",
		"example.com/m/a.go:14.2,15.9 2 0",
		"",
		"a line this tool does not understand",
		"example.com/m/b.go:1.1,1.2 notanumber 0",
	}, "\n"))

	blocks, mode, err := readProfile(path)
	if err != nil {
		t.Fatal(err)
	}

	if mode != "atomic" {
		t.Fatalf("mode = %q", mode)
	}
	// Two good lines kept, the blank, the nonsense and the unparseable one
	// dropped: a profile is machine output, and a tool that panics on one odd
	// line is a tool that stops being run.
	if len(blocks) != 2 {
		t.Fatalf("blocks = %d, want 2", len(blocks))
	}
	if blocks[0].statements != 4 || blocks[0].count != 1 || blocks[0].span != "10.20,12.3" {
		t.Fatalf("first block = %+v", blocks[0])
	}
}

func TestWriteProfileDropsGeneratedAndKeepsSpans(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "reports", "authored.out")

	err := writeProfile(path, "atomic", []block{
		{file: "internal/gateway/settings.go", span: "1.1,2.2", statements: 3, count: 1},
		{file: "internal/api/grpc/ft12/v1/gateway.pb.go", span: "1.1,2.2", statements: 900, count: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)

	if !strings.Contains(got, "mode: atomic") {
		t.Fatalf("profile = %q, want a mode line", got)
	}
	// The span survives, so a coverage viewer can still find the source it
	// refers to rather than being pointed at line one of everything.
	if !strings.Contains(got, "internal/gateway/settings.go:1.1,2.2 3 1") {
		t.Fatalf("profile = %q", got)
	}
	if strings.Contains(got, ".pb.go") {
		t.Fatalf("generated code must not reach the filtered profile: %q", got)
	}
}

func TestPackageOfStripsTheModulePath(t *testing.T) {
	if got := packageOf(modulePath + "/internal/gateway/settings.go"); got != "internal/gateway" {
		t.Fatalf("packageOf() = %q", got)
	}
}

// ---------------------------------------------------------------- ignored

func TestIsSourceLooksAtTheExtension(t *testing.T) {
	for _, name := range []string{"coverage.go", "a/b/route.ts", "deploy.yml", "notes.md"} {
		if !isSource(name) {
			t.Errorf("isSource(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"go-coverage.out", "app.exe", "profile.cov", "image.png"} {
		if isSource(name) {
			t.Errorf("isSource(%q) = true, want false", name)
		}
	}
}

func TestFirstSourceInStopsAtBuildDirectories(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "node_modules", "left-pad", "index.js"), "module.exports = 0")

	found, err := firstSourceIn(root)
	if err != nil {
		t.Fatal(err)
	}
	if found != "" {
		t.Errorf("firstSourceIn = %q, want nothing: a dependency tree is not this repository's source", found)
	}

	mustWrite(t, filepath.Join(root, "pkg", "service.go"), "package pkg")
	found, err = firstSourceIn(root)
	if err != nil {
		t.Fatal(err)
	}
	if found != "service.go" {
		t.Errorf("firstSourceIn = %q, want service.go", found)
	}
}

// TestIgnoredCatchesAPatternThatSwallowsSource reproduces the failure this
// check was written for: `coverage.*`, meant for coverage artefacts, also
// matches coverage.go, and git then hides the file from `git status` and from
// `git add -A` without a word.
func TestIgnoredCatchesAPatternThatSwallowsSource(t *testing.T) {
	root := newGitRepo(t)
	mustWrite(t, filepath.Join(root, ".gitignore"), "coverage.*\n")
	mustWrite(t, filepath.Join(root, "tools", "coverage.go"), "package tools")

	err := runIgnored(root, nil)
	if err == nil {
		t.Fatal("runIgnored() = nil, want a failure naming the hidden file")
	}
	if !strings.Contains(err.Error(), "tools/coverage.go") {
		t.Errorf("error does not name the file: %v", err)
	}
	if !strings.Contains(err.Error(), ".gitignore:1") {
		t.Errorf("error does not name the pattern's line, which is the fix: %v", err)
	}
}

func TestIgnoredAllowsArtefactsAndWholeBuildDirectories(t *testing.T) {
	root := newGitRepo(t)
	mustWrite(t, filepath.Join(root, ".gitignore"), "runtime/\nnode_modules/\n*.out\n")
	mustWrite(t, filepath.Join(root, "runtime", "reports", "go-coverage.out"), "mode: set")
	mustWrite(t, filepath.Join(root, "node_modules", "left-pad", "index.js"), "module.exports = 0")
	mustWrite(t, filepath.Join(root, "kept.go"), "package kept")

	if err := runIgnored(root, nil); err != nil {
		t.Fatalf("runIgnored() = %v, want nil: none of these is hidden source", err)
	}
}

func TestIgnoredReportsAWholeDirectoryOfHiddenSource(t *testing.T) {
	root := newGitRepo(t)
	mustWrite(t, filepath.Join(root, ".gitignore"), "internal/\n")
	mustWrite(t, filepath.Join(root, "internal", "gateway", "poll.go"), "package gateway")

	err := runIgnored(root, nil)
	if err == nil {
		t.Fatal("runIgnored() = nil, want a failure: an ignored directory of source is worse than one file")
	}
	if !strings.Contains(err.Error(), "internal/") {
		t.Errorf("error does not name the directory: %v", err)
	}
}

func newGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	root := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return root
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
