package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const modulePath = "github.com/dimbo1324/ttron-ttr20-time-r"

// boundary is one dependency rule: nothing under Scope may reach anything
// under Forbidden.
//
// A package that is itself under Forbidden is exempt, which is what lets the
// "no legacy" rule be written once over the whole module instead of once per
// top-level directory.
type boundary struct {
	scope     string
	forbidden string
	why       string
}

// The rules, in the order a reader meets them in docs/architecture.md.
//
// The protocol package is the one worth protecting hardest: it is the part
// anyone can lift into another program, and every entry below is a way of
// tying it to this one.
var boundaries = []boundary{
	{modulePath + "/internal/protocol", "net", "the protocol must decode bytes, not sockets"},
	{modulePath + "/internal/protocol", "net/http", "the protocol must decode bytes, not sockets"},
	{modulePath + "/internal/protocol", "google.golang.org/grpc", "the protocol must not know about a control plane"},
	{modulePath + "/internal/protocol", modulePath + "/internal/config", "the protocol must not read the process's configuration"},
	{modulePath + "/internal/protocol", modulePath + "/internal/logging", "the protocol must return errors, not write logs"},
	{modulePath + "/internal/protocol", modulePath + "/internal/platform/logging", "the protocol must return errors, not write logs"},
	{modulePath + "/internal/protocol", modulePath + "/internal/emulator", "the protocol must not depend on either side that speaks it"},
	{modulePath + "/internal/protocol", modulePath + "/internal/gateway", "the protocol must not depend on either side that speaks it"},
	{modulePath + "/internal/protocol", modulePath + "/internal/transport", "the protocol must not depend on how bytes arrive"},
	{modulePath + "/internal/protocol", modulePath + "/internal/api", "the protocol must not depend on how it is served"},
	{modulePath + "/internal/protocol", modulePath + "/internal/app", "the protocol must not depend on process wiring"},
	{modulePath + "/internal/protocol", modulePath + "/internal/adapters", "the protocol must not depend on process wiring"},

	{modulePath + "/internal/emulator", modulePath + "/internal/gateway", "the two sides of the line must stay independent"},
	{modulePath + "/internal/gateway", modulePath + "/internal/emulator", "the two sides of the line must stay independent"},

	{modulePath, modulePath + "/legacy", "active code must not depend on the preserved implementations"},
}

// goPackage is the slice of `go list -json` this check needs.
type goPackage struct {
	ImportPath   string
	Dir          string
	GoFiles      []string
	Imports      []string
	Deps         []string
	TestImports  []string
	XTestImports []string
}

func runArchitecture(root string, _ []string) error {
	packages, err := listPackages(root)
	if err != nil {
		return err
	}

	var f failures
	for _, rule := range boundaries {
		checkBoundary(root, packages, rule, &f)
	}
	checkGeneratedMarkers(root, &f)

	if err := f.err("dependency boundaries broken:"); err != nil {
		return err
	}
	ok("architecture check passed (%d packages, %d boundaries)", len(packages), len(boundaries))
	return nil
}

// listPackages asks the Go toolchain what actually imports what.
//
// `-e` so a package that does not build is reported rather than aborting the
// run: an unresolvable import of a removed package is precisely the kind of
// thing this check should be able to describe.
func listPackages(root string) (map[string]*goPackage, error) {
	cmd := exec.Command("go", "list", "-e", "-json", "./...")
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	packages := make(map[string]*goPackage)
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		pkg := new(goPackage)
		if err := decoder.Decode(pkg); err != nil {
			return nil, fmt.Errorf("go list output: %w", err)
		}
		packages[pkg.ImportPath] = pkg
	}
	return packages, nil
}

func checkBoundary(root string, packages map[string]*goPackage, rule boundary, f *failures) {
	for _, pkg := range packages {
		if !underPath(pkg.ImportPath, rule.scope) || underPath(pkg.ImportPath, rule.forbidden) {
			continue
		}

		// Production code is checked transitively -- reaching `net` through
		// two hops is the same dependency as reaching it directly, and the
		// grep this replaced could not see it.
		if reaches(pkg.Deps, rule.forbidden) {
			path := importChain(packages, pkg.ImportPath, rule.forbidden)
			f.addf("%s\n    %s\n    via %s", describe(pkg.ImportPath, rule), position(root, packages, path), strings.Join(short(path), " -> "))
			continue
		}

		// Test code is checked directly only. A test may legitimately reach
		// far for a fixture; what it must not do is name the forbidden
		// package itself, because then the package cannot be tested without
		// the thing the rule exists to keep out.
		for _, imported := range append(append([]string{}, pkg.TestImports...), pkg.XTestImports...) {
			if underPath(imported, rule.forbidden) {
				f.addf("%s\n    in tests of %s", describe(pkg.ImportPath, rule), shortPath(pkg.ImportPath))
				break
			}
		}
	}
}

func describe(importPath string, rule boundary) string {
	return fmt.Sprintf("%s imports %s -- %s", shortPath(importPath), shortPath(rule.forbidden), rule.why)
}

// underPath reports whether an import path is the given path or inside it.
// Prefix matching alone would make internal/apiserver look like internal/api.
func underPath(importPath, prefix string) bool {
	return importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
}

func reaches(deps []string, forbidden string) bool {
	for _, dep := range deps {
		if underPath(dep, forbidden) {
			return true
		}
	}
	return false
}

// importChain finds the shortest way from a package to the forbidden one.
//
// Breadth-first, so the chain reported is the most direct route rather than
// whichever one the walk happened to reach first -- the difference between
// "you imported this" and a six-hop tour of the module.
func importChain(packages map[string]*goPackage, from, forbidden string) []string {
	type step struct {
		path string
		via  []string
	}

	seen := map[string]bool{from: true}
	queue := []step{{path: from, via: []string{from}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		pkg := packages[current.path]
		if pkg == nil {
			continue
		}
		for _, imported := range pkg.Imports {
			if seen[imported] {
				continue
			}
			seen[imported] = true

			chain := append(append([]string{}, current.via...), imported)
			if underPath(imported, forbidden) {
				return chain
			}
			queue = append(queue, step{path: imported, via: chain})
		}
	}
	return []string{from, forbidden}
}

// position finds where the first hop of a chain is written, so the report ends
// at a file and a line rather than at a package name.
func position(root string, packages map[string]*goPackage, chain []string) string {
	if len(chain) < 2 {
		return ""
	}
	pkg := packages[chain[0]]
	if pkg == nil {
		return ""
	}

	fset := token.NewFileSet()
	for _, name := range pkg.GoFiles {
		full := filepath.Join(pkg.Dir, name)
		file, err := parser.ParseFile(fset, full, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, spec := range file.Imports {
			if strings.Trim(spec.Path.Value, `"`) != chain[1] {
				continue
			}
			return fmt.Sprintf("%s:%d", rel(root, full), fset.Position(spec.Path.Pos()).Line)
		}
	}
	return ""
}

func shortPath(importPath string) string {
	return strings.TrimPrefix(strings.TrimPrefix(importPath, modulePath), "/")
}

func short(chain []string) []string {
	out := make([]string, 0, len(chain))
	for _, item := range chain {
		out = append(out, shortPath(item))
	}
	return out
}

// checkGeneratedMarkers keeps hand edits out of the protobuf output. A file
// without the marker is one somebody has started maintaining by hand, which
// the next `make proto` will silently undo.
func checkGeneratedMarkers(root string, f *failures) {
	dir := filepath.Join(root, "internal", "api", "grpc", "ft12", "v1")
	entries, err := os.ReadDir(dir)
	if err != nil {
		f.addf("cannot read the generated protobuf directory: %v", err)
		return
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".pb.go") {
			continue
		}
		full := filepath.Join(dir, entry.Name())
		body, err := os.ReadFile(full)
		if err != nil {
			f.addf("%s: %v", rel(root, full), err)
			continue
		}
		if !strings.Contains(string(body), "Code generated") {
			f.addf("%s has no \"Code generated\" marker -- is it still generated?", rel(root, full))
		}
	}
}
