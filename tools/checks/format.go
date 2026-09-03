package main

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runFormat checks that every Go file this repository tracks is gofmt-clean.
//
// `go/format` rather than shelling out to the gofmt binary: it is the same
// implementation, it needs nothing on PATH beyond the toolchain already
// running this, and it lets the comparison happen in memory.
func runFormat(root string, _ []string) error {
	files, err := trackedFiles(root, "*.go")
	if err != nil {
		return err
	}

	var f failures
	checked := 0
	for _, file := range files {
		// The preserved implementations under legacy/ are a record of what
		// was, not code anyone maintains; reformatting them would be an edit
		// to a historical document.
		if strings.HasPrefix(file, "legacy/") {
			continue
		}

		full := filepath.Join(root, filepath.FromSlash(file))
		original, err := os.ReadFile(full)
		if err != nil {
			f.addf("%s: %v", file, err)
			continue
		}
		formatted, err := format.Source(original)
		if err != nil {
			f.addf("%s: %v", file, err)
			continue
		}
		checked++
		// Line endings are normalised on both sides. A Windows checkout with
		// autocrlf on holds CRLF on disk and gofmt emits LF, so comparing the
		// bytes as they are would fail every file on one platform and none on
		// the other.
		if normalizeNewlines(string(original)) != normalizeNewlines(string(formatted)) {
			f.addf("%s", file)
		}
	}

	if err := f.err("these files are not gofmt-clean (run `go fmt ./...`):"); err != nil {
		return err
	}
	ok("format check passed (%d files)", checked)
	return nil
}

func normalizeNewlines(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
}

// trackedFiles asks git what belongs to the repository.
//
// Walking the tree instead would sweep up build output, vendored copies and
// whatever an editor left behind; the set worth checking is the set that is
// committed.
func trackedFiles(root string, patterns ...string) ([]string, error) {
	cmd := exec.Command("git", append([]string{"ls-files", "-z"}, patterns...)...)
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}

	var files []string
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" {
			files = append(files, name)
		}
	}
	return files, nil
}
