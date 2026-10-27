// Command testshard splits the top-level tests of one Go package into
// deterministic CI shards and refuses to hand out a partition that does not
// cover every test exactly once.
//
//	go test -list . ./internal/cli | testshard plan -total 5 -index 2 \
//	    -lane unit -skip "$HEAVY" -names names.txt -manifest manifest.txt
//
// plan prints the `go test -run` expression for one shard. It checks the whole
// partition on every call, so each shard job re-proves the cover before it
// runs. verify then reads the manifests the shards left behind and proves the
// cover again from the outside once every shard has finished:
//
//	testshard verify -lanes unit,integration manifest-0.txt manifest-1.txt ...
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// defaultMaxExprBytes keeps the -run expression well below the 128 KiB limit
// Linux puts on a single argument or environment string.
const defaultMaxExprBytes = 100000

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: testshard plan|verify [flags]")
		return 2
	}
	var err error
	switch args[0] {
	case "plan":
		err = runPlan(args[1:], stdin, stdout, stderr)
	case "verify":
		err = runVerify(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "testshard: unknown command %q\n", args[0])
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "testshard: %v\n", err)
		return 1
	}
	return 0
}

func runPlan(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	total := flags.Int("total", 0, "number of shards")
	index := flags.Int("index", -1, "zero-based shard to print")
	lane := flags.String("lane", "default", "lane label recorded in the manifest")
	skip := flags.String("skip", "", "go test -skip expression the shard lane also applies")
	maxBytes := flags.Int("max-expr-bytes", defaultMaxExprBytes, "largest allowed -run expression")
	manifestPath := flags.String("manifest", "", "write this shard's manifest here")
	namesPath := flags.String("names", "", "write this shard's test names here, one per line")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *total < 1 || *index < 0 || *index >= *total {
		return fmt.Errorf("shard index %d is outside 0..%d", *index, *total-1)
	}
	var skipRE *regexp.Regexp
	if *skip != "" {
		var err error
		if skipRE, err = regexp.Compile(*skip); err != nil {
			return fmt.Errorf("-skip does not compile: %w", err)
		}
	}
	names, err := parseList(stdin)
	if err != nil {
		return err
	}
	p, err := newPlan(names, *total, *maxBytes)
	if err != nil {
		return err
	}
	if err := p.verify(); err != nil {
		return err
	}
	for i := range p.shards {
		if p.effective(i, skipRE) == 0 {
			return fmt.Errorf("shard %d of %d has nothing left to run after -skip", i, *total)
		}
	}
	m := manifest{
		Lane: *lane, Shard: *index, Total: *total, Tests: len(p.names),
		Assigned: len(p.shards[*index]), Effective: p.effective(*index, skipRE),
		Universe: p.universeDigest(),
	}
	if *manifestPath != "" {
		if err := os.WriteFile(*manifestPath, []byte(m.String()), 0o644); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}
	}
	if *namesPath != "" {
		body := strings.Join(p.shards[*index], "\n") + "\n"
		if err := os.WriteFile(*namesPath, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write names: %w", err)
		}
	}
	sizes := make([]string, *total)
	for i, owned := range p.shards {
		sizes[i] = fmt.Sprint(len(owned))
	}
	fmt.Fprintf(stderr, "testshard: lane %s shard %d/%d owns %d of %d tests (%d after -skip), -run expression %d bytes, shard sizes %s\n",
		*lane, *index+1, *total, m.Assigned, m.Tests, m.Effective, len(p.exprs[*index]), strings.Join(sizes, " "))
	fmt.Fprintln(stdout, p.exprs[*index])
	return nil
}

func runVerify(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	lanes := flags.String("lanes", "", "comma-separated lanes that must be present")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 0 {
		return fmt.Errorf("verify needs at least one manifest file")
	}
	var manifests []manifest
	for _, path := range flags.Args() {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		m, err := parseManifest(file)
		_ = file.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		manifests = append(manifests, m)
	}
	var required []string
	if *lanes != "" {
		required = strings.Split(*lanes, ",")
	}
	if err := verifyManifests(manifests, required); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "testshard: %d shard manifests cover their packages exactly\n", len(manifests))
	return nil
}
