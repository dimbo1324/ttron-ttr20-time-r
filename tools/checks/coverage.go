package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// generated marks the files nobody writes.
//
// `go test ./... -coverprofile` measures every statement in the module,
// including the ~1,150 in the protobuf output -- close to a quarter of this
// profile. Leaving them in produces a number that moves when the schema
// changes and not when the tests do, which reports on the wrong thing.
func generated(file string) bool {
	base := filepath.Base(file)
	return strings.HasSuffix(base, ".pb.go") || strings.HasSuffix(base, "_grpc.pb.go")
}

// block is one entry of a coverage profile: a span of source, how many
// statements are in it, and how many times the tests ran it.
//
// `span` is kept because a profile can name the same one more than once.
// Running the suite with -coverpkg=./... makes every test binary report on
// every package, so each block appears once per binary; summing them counts a
// module of 4,700 statements as one of 122,000 and reports six percent
// coverage of a well-tested tree.
type block struct {
	file       string
	span       string
	statements int
	count      int
}

type coverage struct {
	total   int
	covered int
}

func (c coverage) percent() float64 {
	if c.total == 0 {
		return 0
	}
	return 100 * float64(c.covered) / float64(c.total)
}

func runCoverage(root string, args []string) error {
	fs := flag.NewFlagSet("coverage", flag.ContinueOnError)
	profile := fs.String("profile", filepath.Join("runtime", "reports", "go-coverage.out"), "coverage profile to read")
	min := fs.Float64("min", 0, "fail below this percentage of covered statements (0 disables)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	blocks, mode, err := readProfile(filepath.Join(root, *profile))
	if err != nil {
		return err
	}
	blocks = merge(blocks)

	var all, authored coverage
	byPackage := map[string]*coverage{}

	for _, b := range blocks {
		all.total += b.statements
		if b.count > 0 {
			all.covered += b.statements
		}
		if generated(b.file) {
			continue
		}
		authored.total += b.statements
		if b.count > 0 {
			authored.covered += b.statements
		}

		pkg := packageOf(b.file)
		if byPackage[pkg] == nil {
			byPackage[pkg] = &coverage{}
		}
		byPackage[pkg].total += b.statements
		if b.count > 0 {
			byPackage[pkg].covered += b.statements
		}
	}

	// The filtered profile is written beside the raw one so a coverage viewer
	// shows the same set of files this summary counted.
	filtered := strings.TrimSuffix(*profile, filepath.Ext(*profile)) + ".authored.out"
	if err := writeProfile(filepath.Join(root, filtered), mode, blocks); err != nil {
		return err
	}

	ok("coverage, hand-written code: %.1f%% of %d statements", authored.percent(), authored.total)
	ok("  including generated:       %.1f%% of %d statements", all.percent(), all.total)
	ok("  filtered profile:          %s", filepath.ToSlash(filtered))

	printWeakest(byPackage)

	if *min > 0 && authored.percent() < *min {
		return fmt.Errorf("coverage %.1f%% is below the %.1f%% floor", authored.percent(), *min)
	}
	return nil
}

// printWeakest names the packages worth looking at, rather than every package.
// A wall of numbers is skimmed; five lines are read.
func printWeakest(byPackage map[string]*coverage) {
	type row struct {
		pkg string
		coverage
	}
	rows := make([]row, 0, len(byPackage))
	for pkg, c := range byPackage {
		// A package with a handful of statements sits at 0% or 100% for
		// reasons that say nothing about how well it is tested.
		if c.total < 10 {
			continue
		}
		rows = append(rows, row{pkg: pkg, coverage: *c})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].percent() < rows[j].percent() })

	limit := min(5, len(rows))
	if limit == 0 {
		return
	}
	ok("  least covered:")
	for _, r := range rows[:limit] {
		ok("    %6.1f%%  %s", r.percent(), r.pkg)
	}
}

func readProfile(path string) ([]block, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("%w (run `make reports` first)", err)
	}
	defer func() { _ = file.Close() }()

	var blocks []block
	mode := "set"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if after, found := strings.CutPrefix(line, "mode: "); found {
			mode = after
			continue
		}
		if line == "" {
			continue
		}

		// name.go:start.col,end.col numStmt count
		space := strings.LastIndexByte(line, ' ')
		if space < 0 {
			continue
		}
		count, err := strconv.Atoi(line[space+1:])
		if err != nil {
			continue
		}
		rest := line[:space]
		space = strings.LastIndexByte(rest, ' ')
		if space < 0 {
			continue
		}
		statements, err := strconv.Atoi(rest[space+1:])
		if err != nil {
			continue
		}
		name, span := rest[:space], ""
		if colon := strings.LastIndexByte(name, ':'); colon >= 0 {
			name, span = name[:colon], name[colon+1:]
		}
		blocks = append(blocks, block{file: name, span: span, statements: statements, count: count})
	}
	return blocks, mode, scanner.Err()
}

// merge folds repeated reports of one span into a single block, keeping the
// highest count: a statement covered by any test binary is covered.
func merge(blocks []block) []block {
	seen := make(map[string]int, len(blocks))
	out := make([]block, 0, len(blocks))

	for _, b := range blocks {
		key := b.file + ":" + b.span
		if index, found := seen[key]; found {
			if b.count > out[index].count {
				out[index].count = b.count
			}
			continue
		}
		seen[key] = len(out)
		out = append(out, b)
	}
	return out
}

func writeProfile(path, mode string, blocks []block) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Rebuilt from the parsed blocks rather than copied line by line, so the
	// filtered file cannot carry a line this tool did not understand.
	var out strings.Builder
	fmt.Fprintf(&out, "mode: %s\n", mode)
	for _, b := range blocks {
		if generated(b.file) {
			continue
		}
		fmt.Fprintf(&out, "%s:%s %d %d\n", b.file, b.span, b.statements, b.count)
	}
	return os.WriteFile(path, []byte(out.String()), 0o644)
}

func packageOf(file string) string {
	dir := filepath.ToSlash(filepath.Dir(file))
	return strings.TrimPrefix(strings.TrimPrefix(dir, modulePath), "/")
}
