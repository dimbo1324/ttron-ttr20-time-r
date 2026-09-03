package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// stage is one step of the release gate: a name a reader will see when it
// fails, and the command behind it.
type stage struct {
	name string
	argv []string
	// optional marks a stage that is skipped when its tool is missing rather
	// than failing the run.
	optional bool
	// quiet drops the stage's stdout. For a stage whose only answer is its
	// exit code -- `compose config` prints the entire resolved file -- the
	// output buries every other line of the run. Stderr always comes through,
	// so a failure still explains itself.
	quiet bool
}

// runRelease is the gate before a tag: everything the other checks cover, plus
// the things only a full build can tell you.
//
// The order is cheapest-first. A missing gofmt is a five-second answer and a
// compose build is a five-minute one, and there is no sense paying for the
// second to be told about the first.
func runRelease(root string, _ []string) error {
	self := []string{"go", "run", "./tools/checks"}

	stages := []stage{
		{name: "format", argv: append(append([]string{}, self...), "format")},
		{name: "architecture", argv: append(append([]string{}, self...), "architecture")},
		{name: "doc links", argv: append(append([]string{}, self...), "doc-links")},
		{name: "vet", argv: []string{"go", "vet", "./..."}},
		{name: "tests", argv: []string{"go", "test", "./..."}},
		{name: "build", argv: []string{"go", "build", "./..."}},
		{name: "compose config", argv: []string{"docker", "compose", "config"}, optional: true, quiet: true},
		{name: "compose config (observability)", argv: []string{"docker", "compose", "--profile", "observability", "config"}, optional: true, quiet: true},
		{name: "cleanup dry-run", argv: append(append([]string{}, self...), "clean-runtime", "--dry-run")},
	}

	for _, s := range stages {
		if s.optional {
			if _, err := exec.LookPath(s.argv[0]); err != nil {
				ok("  skip  %-30s (%s is not installed)", s.name, s.argv[0])
				continue
			}
		}
		ok("  run   %-30s %s", s.name, strings.Join(s.argv, " "))
		if err := run(root, s.argv, s.quiet); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}

	ok("release check passed")
	return nil
}

// run executes a stage, letting its output through so a failure is readable
// where it happened rather than summarised after the fact.
func run(root string, argv []string, quiet bool) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	if !quiet {
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()
}
