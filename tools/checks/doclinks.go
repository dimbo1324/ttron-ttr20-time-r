package main

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// markdownLink matches both links and images; the leading `!` is allowed
// because a broken image is a broken link with a different symptom.
var markdownLink = regexp.MustCompile(`!?\[[^\]]*\]\(([^)]+)\)`)

// runDocLinks checks that every local Markdown link points at something.
//
// Only local ones. Reaching out to every external URL would make the check
// slow, flaky and dependent on the network, and would fail a pull request
// because somebody else's site was down.
func runDocLinks(root string, _ []string) error {
	files, err := trackedFiles(root, "*.md")
	if err != nil {
		return err
	}

	var f failures
	links := 0
	for _, file := range files {
		full := filepath.Join(root, filepath.FromSlash(file))
		body, err := os.ReadFile(full)
		if err != nil {
			f.addf("%s: %v", file, err)
			continue
		}
		for _, target := range localTargets(string(body)) {
			links++
			if !resolves(root, file, target) {
				f.addf("%s -> %s", file, target)
			}
		}
	}

	if err := f.err("these local Markdown links point at nothing:"); err != nil {
		return err
	}
	ok("doc link check passed (%d files, %d local links)", len(files), links)
	return nil
}

// localTargets pulls the link targets worth checking out of one document.
func localTargets(body string) []string {
	var targets []string
	for _, match := range markdownLink.FindAllStringSubmatch(body, -1) {
		target := strings.TrimSpace(match[1])

		// Angle brackets are how Markdown carries a path with a space in it,
		// and they settle the question the next rule would otherwise get
		// wrong: inside them, whitespace belongs to the path.
		//
		// The scripts this replaced trimmed the brackets first and cut at the
		// first space afterwards, so `[x](<a file.md>)` was checked as `a` --
		// a target that resolves to nothing and was never reported, because
		// it was not the target written on the page.
		if strings.HasPrefix(target, "<") && strings.HasSuffix(target, ">") {
			target = target[1 : len(target)-1]
		} else if cut := strings.IndexAny(target, " \t"); cut >= 0 {
			// A title after the target: `[text](path "Title")`.
			target = target[:cut]
		}
		// Anchors are checked by the reader, not by this: resolving them
		// would mean modelling how every Markdown renderer slugifies a
		// heading, and getting that subtly wrong is worse than not trying.
		if cut := strings.Index(target, "#"); cut >= 0 {
			target = target[:cut]
		}
		if target == "" || isExternal(target) {
			continue
		}
		targets = append(targets, target)
	}
	return targets
}

func isExternal(target string) bool {
	for _, scheme := range []string{"http://", "https://", "mailto:", "tel:"} {
		if strings.HasPrefix(target, scheme) {
			return true
		}
	}
	return false
}

// resolves reports whether a target exists on disk. A leading slash is read as
// repository-relative, which is how these documents are written and read on
// GitHub, rather than as the filesystem root.
func resolves(root, from, target string) bool {
	decoded, err := url.PathUnescape(target)
	if err != nil {
		decoded = target
	}

	var candidate string
	if strings.HasPrefix(decoded, "/") {
		candidate = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(decoded, "/")))
	} else {
		candidate = filepath.Join(root, filepath.FromSlash(path.Dir(from)), filepath.FromSlash(decoded))
	}

	_, err = os.Stat(candidate)
	return err == nil
}
