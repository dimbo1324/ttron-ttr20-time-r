// Command checks runs the repository's own rules -- the ones the compiler and
// the test suite cannot express.
//
// It replaces five pairs of shell and PowerShell scripts. Those pairs were two
// independent implementations of one rule each, which is exactly the shape
// that drifts: CI had quietly settled on the PowerShell half for two of them,
// leaving the shell half unrun and, on a Windows checkout, unrunnable.
//
// Go rather than Python because this is a Go repository: the toolchain is
// already required and its version is already pinned in go.mod, so the checks
// cost no new dependency and run identically wherever `go` does. It also buys
// a real import graph for the architecture check instead of a grep.
//
// Usage:
//
//	go run ./tools/checks <name> [flags]
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type check struct {
	name    string
	summary string
	run     func(root string, args []string) error
}

var checks = []check{
	{"architecture", "dependency boundaries between packages", runArchitecture},
	{"format", "gofmt over the Go files this repository tracks", runFormat},
	{"doc-links", "local Markdown links resolve to something", runDocLinks},
	{"clean-runtime", "remove build and runtime artefacts [--dry-run]", runCleanRuntime},
	{"release", "everything above, plus tests, build and compose config", runRelease},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	name := os.Args[1]
	for _, c := range checks {
		if c.name != name {
			continue
		}
		root, err := repoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "checks: %v\n", err)
			os.Exit(1)
		}
		if err := c.run(root, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "%s check failed: %v\n", name, err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "checks: unknown check %q\n\n", name)
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./tools/checks <check> [flags]")
	fmt.Fprintln(os.Stderr)
	for _, c := range checks {
		fmt.Fprintf(os.Stderr, "  %-14s %s\n", c.name, c.summary)
	}
}

// repoRoot walks up from the working directory to the module root, so a check
// behaves the same whether it is run from the repository root, from an editor
// rooted somewhere else, or from a Makefile in a subdirectory.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod found in this directory or any parent")
		}
		dir = parent
	}
}

// failures collects everything wrong before saying anything.
//
// A check that stops at the first problem makes a reader fix one line, run it
// again, and find the next -- which is the slowest possible way to learn that
// eleven files need formatting.
type failures struct {
	items []string
}

func (f *failures) addf(format string, args ...any) {
	f.items = append(f.items, fmt.Sprintf(format, args...))
}

// err renders the collected failures as one error, or nil when there are none.
func (f *failures) err(headline string) error {
	if len(f.items) == 0 {
		return nil
	}
	sort.Strings(f.items)
	return fmt.Errorf("%s\n  %s", headline, strings.Join(f.items, "\n  "))
}

func ok(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

/*
  Paths are reported relative to the repository root throughout: an absolute
  path from a CI runner is noise, and a relative one is what a reader can paste
  into an editor.
*/

func rel(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}
