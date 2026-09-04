package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runIgnored looks for source that git has been told to ignore.
//
// This check exists because of a specific afternoon. `.gitignore` carried
// `coverage.*` from the first commit, meaning coverage artefacts. A later
// commit added `tools/checks/coverage.go`, and the pattern swallowed it: the
// file never appeared in `git status`, `git add -A` skipped it without a word,
// and every check passed locally because locally the file was there. CI cloned
// the repository, found `undefined: runCoverage`, and failed eight jobs.
//
// That is the shape worth guarding against. A too-broad ignore pattern does
// not fail on the machine that has the file. It fails on every machine that
// does not, which is all of them.
func runIgnored(root string, _ []string) error {
	entries, err := listIgnored(root)
	if err != nil {
		return err
	}

	var f failures
	for _, entry := range entries {
		if strings.HasSuffix(entry, "/") {
			// An ignored directory is usually deliberate -- node_modules,
			// runtime, bin. One holding source is not.
			name := lastSegment(strings.TrimSuffix(entry, "/"))
			if buildDirs[name] {
				continue
			}
			if found, err := firstSourceIn(filepath.Join(root, entry)); err != nil {
				return err
			} else if found != "" {
				f.addf("%s is ignored and holds source (%s)", entry, found)
			}
			continue
		}
		if !isSource(entry) || generatedAndIgnored[entry] {
			continue
		}
		f.addf("%s is ignored by %s", entry, ignoreReason(root, entry))
	}

	headline := "source hidden from git; narrow the pattern, or record the exception in tools/checks/ignored.go:"
	if err := f.err(headline); err != nil {
		return err
	}
	ok("ignore check passed (%d ignored paths, none of them source)", len(entries))
	return nil
}

// generatedAndIgnored is the short list of files that are source-shaped, are
// ignored, and should be: something else rewrites them on every build.
var generatedAndIgnored = map[string]bool{
	// Next rewrites this on every `next dev` and `next build`, and its own
	// template ignores it.
	"web/next-env.d.ts": true,
}

// buildDirs are directories whose whole contents are ignored on purpose.
var buildDirs = map[string]bool{
	".frames":      true,
	".next":        true,
	"bin":          true,
	"coverage":     true,
	"dist":         true,
	"node_modules": true,
	"runtime":      true,
	"tmp":          true,
	"vendor":       true,
}

var sourceExtensions = map[string]bool{
	".go": true, ".proto": true,
	".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
	".css": true, ".sh": true, ".ps1": true, ".py": true, ".sql": true,
	".yml": true, ".yaml": true, ".json": true, ".md": true,
}

func isSource(file string) bool {
	return sourceExtensions[filepath.Ext(file)]
}

func lastSegment(entry string) string {
	segments := strings.Split(entry, "/")
	return segments[len(segments)-1]
}

// firstSourceIn names one source file under dir, or "" if it holds none. One
// example is enough: the point is to say that a directory of source is
// invisible, not to list it.
func firstSourceIn(dir string) (string, error) {
	var found string
	err := filepath.WalkDir(dir, func(name string, entry fs.DirEntry, walkErr error) error {
		// A path that cannot be read is a path this check cannot judge, and
		// one unreadable directory is not a reason to fail the run. SkipDir
		// rather than nil so that the decision is written down instead of
		// being an ignored error.
		if walkErr != nil {
			return fs.SkipDir
		}
		if entry.IsDir() {
			if buildDirs[entry.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if isSource(name) {
			found = filepath.Base(name)
			return fs.SkipAll
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return found, nil
}

func listIgnored(root string) ([]string, error) {
	// --others --ignored is "on disk, untracked, and told to be ignored" --
	// exactly the set this check is about. --directory collapses a wholly
	// ignored directory to its own name rather than listing the ninety
	// thousand files under node_modules.
	out, err := gitOutput(root, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			paths = append(paths, filepath.ToSlash(line))
		}
	}
	return paths, nil
}

// ignoreReason names the file and line of the pattern responsible, so the fix
// is a line number rather than a search through several ignore files.
func ignoreReason(root, file string) string {
	out, err := gitOutput(root, "check-ignore", "--verbose", file)
	if err != nil {
		return "an ignore pattern"
	}
	if fields := strings.SplitN(strings.TrimSpace(out), "\t", 2); len(fields) > 0 && fields[0] != "" {
		return fields[0]
	}
	return "an ignore pattern"
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
