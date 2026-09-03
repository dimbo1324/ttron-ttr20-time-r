package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// artefactDirs are removed whole.
var artefactDirs = []string{"tmp", "runtime", "bin", "dist"}

// artefactFiles are removed by name, at the repository root only.
var artefactFiles = []string{"coverage.out"}

// artefactSuffixes are removed by extension, at the repository root only.
//
// Root only on purpose: `*.log` swept recursively would take a fixture out of
// a testdata directory, and a cleanup that eats test data is worse than no
// cleanup at all.
var artefactSuffixes = []string{".log", ".out", ".tsbuildinfo"}

func runCleanRuntime(root string, args []string) error {
	fs := flag.NewFlagSet("clean-runtime", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "list what would be removed and remove nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	targets, err := artefacts(root)
	if err != nil {
		return err
	}

	removed := 0
	for _, target := range targets {
		// Belt and braces around a recursive delete. The paths are built from
		// the module root and cannot escape it, but this function is one bad
		// edit away from being able to, and the cost of checking is nothing.
		if !within(root, target) {
			return fmt.Errorf("refusing to remove a path outside the repository: %s", target)
		}
		if *dryRun {
			ok("would remove %s", rel(root, target))
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove %s: %w", rel(root, target), err)
		}
		ok("remove %s", rel(root, target))
		removed++
	}

	if *dryRun {
		ok("cleanup dry-run complete (%d artefacts)", len(targets))
		return nil
	}
	ok("cleanup complete (%d artefacts)", removed)
	return nil
}

// artefacts lists what exists right now, so the caller reports only real
// removals rather than every path it might ever have to clean.
func artefacts(root string) ([]string, error) {
	var targets []string

	for _, name := range append(append([]string{}, artefactDirs...), artefactFiles...) {
		full := filepath.Join(root, name)
		if _, err := os.Stat(full); err == nil {
			targets = append(targets, full)
		}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		for _, suffix := range artefactSuffixes {
			if !strings.HasSuffix(entry.Name(), suffix) {
				continue
			}
			full := filepath.Join(root, entry.Name())
			if !contains(targets, full) {
				targets = append(targets, full)
			}
			break
		}
	}
	return targets, nil
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// within reports whether a path is the root or inside it, comparing cleaned
// absolute paths so `..` cannot walk out.
func within(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
