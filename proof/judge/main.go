// Command judge reads the observations run-proof.sh collected and writes
// REPORT.md. It depends on nothing but the standard library and never imports
// go-cty: every expectation is recomputed here from first principles (exhaustive
// substitution for Equals, exact big-integer lengths for SetProduct, rational
// comparison for refinement bounds), and every check is printed as PASS or FAIL
// whatever the outcome.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	resultsDir string
	report     strings.Builder
	checks     []check
)

type check struct {
	Section string
	Name    string
	Pass    bool
	Detail  string
}

func record(section, name string, pass bool, detail string) {
	checks = append(checks, check{section, name, pass, detail})
}

func w(format string, args ...any) {
	fmt.Fprintf(&report, format, args...)
	report.WriteByte('\n')
}

func status(pass bool) string {
	if pass {
		return "✅ PASS"
	}
	return "❌ FAIL"
}

func readJSONL[T any](path string) []T {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var ret []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<30)
	for sc.Scan() {
		var v T
		if err := json.Unmarshal(sc.Bytes(), &v); err != nil {
			panic(fmt.Errorf("%s: %w", path, err))
		}
		ret = append(ret, v)
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	return ret
}

func readText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(b), "\n")
}

const (
	upstreamRelease = "upstream-v1.19.0"
	upstreamMain    = "upstream-main"
	forkEquals      = "fork-equals"
	forkSetProduct  = "fork-setproduct"
	forkBounds      = "fork-bounds"
	forkMain        = "fork-main"
	pyvider         = "pyvider-cty"
	pyviderSection  = "pyvider-cty"
)

var variants = []string{upstreamRelease, upstreamMain, forkEquals, forkSetProduct, forkBounds, forkMain}

// allBuilds is every go-cty build followed by pyvider-cty, which is graded on
// the same cases but has no Go-specific results (random-call replay, memory
// statistics, OpenTofu).
var allBuilds = []string{upstreamRelease, upstreamMain, forkEquals, forkSetProduct, forkBounds, forkMain, pyvider}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: judge RESULTS_DIR REPORT.md")
		os.Exit(2)
	}
	resultsDir = os.Args[1]

	body := func() string {
		provenance()
		branches()
		equalsSection()
		setProductSection()
		boundsSection()
		fullSuites()
		limitations()
		return report.String()
	}()

	report.Reset()
	w("# go-cty correctness fixes — proof report")
	w("")
	w("Generated %s by `run-proof.sh`. Every number, outcome and PASS/FAIL below is computed from the files in `results/`.", time.Now().Format("2006-01-02 15:04 MST"))
	w("")
	w("Three defects in go-cty, each shown to exist in the released and current upstream code, and shown to be fixed on the `livingstaccato/go-cty` fork. The same harness source is built against every variant; a separate judge (standard library only, no go-cty import) computes the expected answers independently and grades every build.")
	w("")
	w("Alongside the Go builds, **pyvider-cty** — the Python port of go-cty — is put through the same cases by a Python harness and graded by the same judge, so every section also shows where the Python implementation stands.")
	w("")
	summary()
	w("## Reproducing this report")
	w("")
	w("```sh")
	w("./run-proof.sh all        # from this directory; or one step: prs | variants | pyvider | unit-tests | tofu-stock | tofu-built | report")
	w("```")
	w("")
	w("Needs git, Go %s or newer, OpenTofu 1.12.6 on `PATH`, uv, jq, and gh logged in (`gh auth login`); `README.md` has the details. Everything else is fetched over https into `build/`.", "1.25")
	w("")
	w("| Path | What it is |")
	w("|---|---|")
	w("| `proof/harness/main.go` | The only program that calls go-cty. Built once per variant; prints observations, contains no expectations. |")
	w("| `proof/pyharness/harness.py` | The same cases against pyvider-cty, printed in the same shapes. Contains no expectations. |")
	w("| `proof/judge/main.go` | Standard library only. Recomputes ground truth, grades each build, writes this file. |")
	w("| `tofu/equals`, `tofu/count` | OpenTofu configurations, planned repeatedly (never applied). |")
	w("| `results/` | Raw observations from the last run. |")
	w("| `build/` | Not committed. Every go-cty build, the fork's clone (`build/go-cty`), OpenTofu v1.12.6 source (`build/opentofu`) and the two OpenTofu builds from it. |")
	w("| `work/` | Not committed. The source trees the regression tests and full suites run in. |")
	w("")
	report.WriteString(body)

	if err := os.WriteFile(os.Args[2], []byte(report.String()), 0o644); err != nil {
		panic(err)
	}
	failed := 0
	for _, c := range checks {
		if !c.Pass {
			failed++
			fmt.Printf("FAIL  [%s] %s — %s\n", c.Section, c.Name, c.Detail)
		}
	}
	fmt.Printf("%d checks, %d failed; wrote %s\n", len(checks), failed, os.Args[2])
	if failed > 0 {
		os.Exit(1)
	}
}

func summary() {
	w("## Verdict")
	w("")
	sections := []string{"Fork branches", "Value.Equals", "SetProductFunc", "Refinement bounds", "Upstream test suites", pyviderSection}
	w("| Area | Checks | Passed | Result |")
	w("|---|---:|---:|---|")
	for _, s := range sections {
		total, passed := 0, 0
		for _, c := range checks {
			if c.Section == s {
				total++
				if c.Pass {
					passed++
				}
			}
		}
		w("| %s | %d | %d | %s |", s, total, passed, status(total > 0 && total == passed))
	}
	w("")
	w("<details><summary>Every check</summary>")
	w("")
	w("| Area | Check | Result | Evidence |")
	w("|---|---|---|---|")
	for _, c := range checks {
		w("| %s | %s | %s | %s |", c.Section, c.Name, status(c.Pass), strings.ReplaceAll(c.Detail, "|", "\\|"))
	}
	w("")
	w("</details>")
	w("")
}

// ---------------------------------------------------------------------------

type variantMeta struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
	Go     string `json:"go"`
}

type moduleInfo struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Replace *struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Replace"`
}

func provenance() {
	w("## Provenance")
	w("")
	w("Each build fetched go-cty through `proxy.golang.org`; the `h1:` hashes are the ones `go mod verify` checked, and for public modules they are also recorded in `sum.golang.org`. Fork branches are resolved to the commit the fork's remote had at run time, so a result cannot come from an unpushed local edit.")
	w("")
	w("| Variant | Source | Branch / ref | Commit | Module actually built | `go.sum` hash | Go |")
	w("|---|---|---|---|---|---|---|")
	for _, v := range variants {
		var meta variantMeta
		if b, err := os.ReadFile(filepath.Join(resultsDir, v, "variant.json")); err == nil {
			_ = json.Unmarshal(b, &meta)
		}
		var mod moduleInfo
		if b, err := os.ReadFile(filepath.Join(resultsDir, v, "module.json")); err == nil {
			_ = json.Unmarshal(b, &mod)
		}
		built := mod.Path + "@" + mod.Version
		if mod.Replace != nil {
			built = mod.Replace.Path + "@" + mod.Replace.Version
		}
		hash := ""
		for _, line := range strings.Split(readText(filepath.Join(resultsDir, v, "go.sum.lines")), "\n") {
			f := strings.Fields(line)
			if len(f) == 3 && !strings.HasSuffix(f[1], "/go.mod") {
				hash = f[2]
			}
		}
		commit := meta.Commit
		if len(commit) > 12 {
			commit = commit[:12]
		}
		w("| `%s` | %s | `%s` | `%s` | `%s` | `%s` | %s |", v, meta.Source, meta.Ref, commit, built, hash, meta.Go)
	}
	w("")
	pyviderProvenance()
}

type pyviderMeta struct {
	Ref       string `json:"ref"`
	Commit    string `json:"commit"`
	Installed string `json:"installed_commit"`
	Version   string `json:"version"`
	Python    string `json:"python"`
}

func loadPyviderMeta() pyviderMeta {
	var meta pyviderMeta
	if b, err := os.ReadFile(filepath.Join(resultsDir, pyvider, "variant.json")); err == nil {
		_ = json.Unmarshal(b, &meta)
	}
	return meta
}

func pyviderProvenance() {
	meta := loadPyviderMeta()
	w("pyvider-cty is installed with `uv pip install` straight from GitHub, at the commit its branch had at run time; the installed commit is read back from the package's own `direct_url.json`.")
	w("")
	w("| Build | Branch | Commit on GitHub | Commit installed | Package | Python |")
	w("|---|---|---|---|---|---|")
	w("| `%s` | `%s` | `%.12s` | `%.12s` | `pyvider-cty %s` | %s |", pyvider, meta.Ref, meta.Commit, meta.Installed, meta.Version, meta.Python)
	w("")
	record(pyviderSection, "installed package is the commit GitHub had at run time", meta.Commit != "" && meta.Commit == meta.Installed, fmt.Sprintf("GitHub %.12s, installed %.12s", meta.Commit, meta.Installed))
}

type pullRequest struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	HeadRefName string `json:"headRefName"`
	HeadRefOid  string `json:"headRefOid"`
	BaseRefName string `json:"baseRefName"`
	State       string `json:"state"`
}

// headParents is one PR head as GitHub reports it: its parents, and whether a
// second parent is already on the base branch, i.e. the head only merged the
// base into the commit before it.
type headParents struct {
	Head               string   `json:"head"`
	Parents            []string `json:"parents"`
	SecondParentOnBase bool     `json:"second_parent_on_base"`
}

func branches() {
	const section = "Fork branches"
	w("## The fork's pull requests carry exactly what was tested")
	w("")
	w("Pull requests as GitHub reports them (`gh pr list -R livingstaccato/go-cty`), against the commit each build above was pinned to for the same branch. A PR head that only merged its base branch into the tested commit also counts: its first parent is the tested commit and its second is already on the base. Each fix is tested at that first parent, so it is measured on its own, without the fixes merged before it.")
	w("")
	var prs []pullRequest
	if b, err := os.ReadFile(filepath.Join(resultsDir, "prs.json")); err == nil {
		_ = json.Unmarshal(b, &prs)
	}
	tested := map[string]string{}
	for _, v := range variants {
		var meta variantMeta
		if b, err := os.ReadFile(filepath.Join(resultsDir, v, "variant.json")); err == nil && json.Unmarshal(b, &meta) == nil && meta.Source == "fork" {
			tested[meta.Ref] = meta.Commit
		}
	}
	parents := map[string]headParents{}
	if b, err := os.ReadFile(filepath.Join(resultsDir, "prs-parents.json")); err == nil {
		var list []headParents
		_ = json.Unmarshal(b, &list)
		for _, p := range list {
			parents[p.Head] = p
		}
	}
	sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
	w("| PR | Title | Branch → base | PR head | Commit tested | Relation | Result |")
	w("|---|---|---|---|---|---|---|")
	for _, pr := range prs {
		commit := tested[pr.HeadRefName]
		p := parents[pr.HeadRefOid]
		relation := "different"
		switch {
		case commit == "":
			relation = "not tested"
		case commit == pr.HeadRefOid:
			relation = "head"
		case len(p.Parents) == 2 && p.Parents[0] == commit && p.SecondParentOnBase:
			relation = "head merges base into it"
		}
		ok := relation == "head" || relation == "head merges base into it"
		w("| [#%d](%s) | %s | `%s` → `%s` | `%.12s` | `%.12s` | %s | %s |", pr.Number, pr.URL, strings.ReplaceAll(pr.Title, "|", "\\|"), pr.HeadRefName, pr.BaseRefName, pr.HeadRefOid, commit, relation, status(ok))
		record(section, fmt.Sprintf("PR #%d head is the tested commit, or only merges its base into it", pr.Number), ok, fmt.Sprintf("PR %.12s, tested %.12s, %s", pr.HeadRefOid, commit, relation))
	}
	if len(prs) == 0 {
		record(section, "pull requests read from GitHub", false, "results/prs.json missing — run `./run-proof.sh prs`")
	}
	if commit := tested["main"]; commit != "" {
		w("| — | all three pull requests merged | `main` | — | `%.12s` | (no PR) | — |", commit)
	}
	w("")
}

// ---------------------------------------------------------------------------
// Value.Equals

type equalsObservation struct {
	Category string `json:"category"`
	GoString string `json:"gostring"`
}

type equalsCase struct {
	Kind     string              `json:"kind"`
	LHS      string              `json:"lhs"`
	RHS      string              `json:"rhs"`
	Marked   bool                `json:"marked"`
	Observed []equalsObservation `json:"observed"`
	Truth    []string            `json:"truth"`
}

func (c equalsCase) key() string {
	return fmt.Sprintf("%s|%s|%s|%t", c.Kind, c.LHS, c.RHS, c.Marked)
}

// Independent ground truth, recomputed here rather than trusted from the
// harness: substitute every combination of p, q and a fresh value r for the
// unknown positions, and collect the equality results that are possible.
func equalsTruth(lhs, rhs string) map[string]bool {
	joined := []byte(lhs + rhs)
	var unknowns []int
	for i, c := range joined {
		if c == 'U' {
			unknowns = append(unknowns, i)
		}
	}
	subs := []byte{'p', 'q', 'r'}
	combos := 1
	for range unknowns {
		combos *= len(subs)
	}
	ret := map[string]bool{}
	for combo := range combos {
		concrete := append([]byte(nil), joined...)
		rest := combo
		for _, pos := range unknowns {
			concrete[pos] = subs[rest%len(subs)]
			rest /= len(subs)
		}
		ret[strconv.FormatBool(string(concrete[:len(lhs)]) == string(concrete[len(lhs):]))] = true
	}
	return ret
}

func sound(cat string, truth map[string]bool) bool {
	switch cat {
	case "unknown":
		return true
	case "true":
		return truth["true"] && !truth["false"]
	case "false":
		return truth["false"] && !truth["true"]
	}
	return false
}

func mostPrecise(truth map[string]bool) string {
	if len(truth) == 1 {
		for k := range truth {
			return k
		}
	}
	return "unknown"
}

type equalsStats struct {
	cases, nondet, unsound, imprecise map[string]int
	byKey                             map[string]equalsCase
	truthMismatch                     int
}

func loadEquals(v string) *equalsStats {
	cases := readJSONL[equalsCase](filepath.Join(resultsDir, v, "equals.jsonl"))
	if cases == nil {
		return nil
	}
	s := &equalsStats{map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]equalsCase{}, 0}
	for _, c := range cases {
		truth := equalsTruth(c.LHS, c.RHS)
		harnessTruth := map[string]bool{}
		for _, t := range c.Truth {
			harnessTruth[t] = true
		}
		if fmt.Sprint(truth) != fmt.Sprint(harnessTruth) {
			s.truthMismatch++
		}
		s.byKey[c.key()] = c
		s.cases[c.Kind]++
		if len(c.Observed) > 1 {
			s.nondet[c.Kind]++
		}
		bad, loose := false, false
		for _, o := range c.Observed {
			if !sound(o.Category, truth) {
				bad = true
			}
			if o.Category != mostPrecise(truth) {
				loose = true
			}
		}
		if bad {
			s.unsound[c.Kind]++
		}
		if loose {
			s.imprecise[c.Kind]++
		}
	}
	return s
}

func observedSet(c equalsCase) map[string]bool {
	ret := map[string]bool{}
	for _, o := range c.Observed {
		ret[o.GoString] = true
	}
	return ret
}

func categories(c equalsCase) string {
	seen := map[string]bool{}
	for _, o := range c.Observed {
		seen[o.Category] = true
	}
	var cats []string
	for k := range seen {
		cats = append(cats, k)
	}
	sort.Strings(cats)
	return strings.Join(cats, " / ")
}

func equalsSection() {
	const section = "Value.Equals"
	w("## 1. `Value.Equals` gives different answers for the same objects and maps")
	w("")
	w("**Claim.** When an object or map holds an unknown member *and* a member that definitely differs, upstream `Equals` answers `false` or unknown depending on Go's randomized map iteration order. The fix makes a definite difference win, so the answer is always `false`.")
	w("")
	w("**Method.** Every pair of operands with 1–4 positions, each position one of `p`, `q` or unknown (`U`), for objects, maps, lists and tuples — 29,520 unmarked pairs, plus 3,276 with a `sensitive` mark on the left operand's first position. Each pair is compared **200 times** in one process. Ground truth is computed by the judge, not by go-cty: every unknown is replaced by each of `p`, `q` and a fresh value `r` in every combination, and the set of possible concrete results is collected. An answer is *sound* if no substitution contradicts it; the *most precise* answer is the concrete result when all substitutions agree, and unknown otherwise.")
	w("")

	stats := map[string]*equalsStats{}
	for _, v := range allBuilds {
		stats[v] = loadEquals(v)
	}
	if stats[upstreamMain] == nil || stats[forkEquals] == nil {
		w("_Equals results missing — run `./run-proof.sh variants`._")
		record(section, "results present", false, "equals.jsonl missing")
		return
	}

	kinds := []string{"object", "map", "list", "tuple"}
	w("### Results per build")
	w("")
	w("| Build | Pairs | Different answers on repeat (object / map / list / tuple) | Unsound answers | Object/map answers less precise than possible |")
	w("|---|---:|---|---:|---:|")
	for _, v := range allBuilds {
		s := stats[v]
		if s == nil {
			w("| `%s` | — | not run | — | — |", v)
			continue
		}
		total, unsound := 0, 0
		var nd []string
		for _, k := range kinds {
			total += s.cases[k]
			unsound += s.unsound[k]
			nd = append(nd, strconv.Itoa(s.nondet[k]))
		}
		w("| `%s` | %d | %s | %d | %d |", v, total, strings.Join(nd, " / "), unsound, s.imprecise["object"]+s.imprecise["map"])
	}
	w("")

	// Checks.
	for _, v := range variants {
		if s := stats[v]; s != nil {
			record(section, fmt.Sprintf("judge's ground truth matches harness's (`%s`)", v), s.truthMismatch == 0, fmt.Sprintf("%d mismatches", s.truthMismatch))
		}
	}
	for _, v := range []string{upstreamRelease, upstreamMain} {
		if s := stats[v]; s != nil {
			record(section, fmt.Sprintf("bug is real: `%s` answers objects inconsistently", v), s.nondet["object"] > 0, fmt.Sprintf("%d object pairs, %d map pairs gave two different answers", s.nondet["object"], s.nondet["map"]))
			record(section, fmt.Sprintf("bug is real: `%s` answers maps inconsistently", v), s.nondet["map"] > 0, fmt.Sprintf("%d map pairs", s.nondet["map"]))
			record(section, fmt.Sprintf("lists and tuples were already consistent on `%s`", v), s.nondet["list"] == 0 && s.nondet["tuple"] == 0, fmt.Sprintf("%d list, %d tuple", s.nondet["list"], s.nondet["tuple"]))
		}
	}
	for _, v := range []string{forkEquals, forkMain} {
		s := stats[v]
		if s == nil {
			record(section, fmt.Sprintf("`%s` results present", v), false, "missing")
			continue
		}
		nd, un := 0, 0
		for _, k := range kinds {
			nd += s.nondet[k]
			un += s.unsound[k]
		}
		record(section, fmt.Sprintf("fixed: `%s` gives one answer for every pair", v), nd == 0, fmt.Sprintf("%d pairs with more than one answer", nd))
		record(section, fmt.Sprintf("fixed: `%s` never gives an unsound answer", v), un == 0, fmt.Sprintf("%d unsound", un))
		record(section, fmt.Sprintf("fixed: `%s` object/map answers are the most precise possible", v), s.imprecise["object"]+s.imprecise["map"] == 0, fmt.Sprintf("%d less precise", s.imprecise["object"]+s.imprecise["map"]))

		invented, listChanged := 0, 0
		var inventedExample string
		for key, fc := range s.byKey {
			uc, ok := stats[upstreamMain].byKey[key]
			if !ok {
				invented++
				continue
			}
			up := observedSet(uc)
			for gs := range observedSet(fc) {
				if !up[gs] {
					invented++
					if inventedExample == "" {
						inventedExample = key + " → " + gs
					}
				}
			}
			if (fc.Kind == "list" || fc.Kind == "tuple") && fmt.Sprint(observedSet(fc)) != fmt.Sprint(up) {
				listChanged++
			}
		}
		record(section, fmt.Sprintf("`%s` never answers anything upstream never answered (marks included)", v), invented == 0, fmt.Sprintf("%d new answers %s", invented, inventedExample))
		record(section, fmt.Sprintf("`%s` leaves every list and tuple answer exactly as upstream", v), listChanged == 0, fmt.Sprintf("%d changed", listChanged))
	}
	for _, v := range []string{forkSetProduct, forkBounds} {
		if s := stats[v]; s != nil && stats[upstreamMain] != nil {
			same := s.nondet["object"] > 0 && s.nondet["map"] > 0
			record(section, fmt.Sprintf("isolation: `%s` (no Equals change) still shows the Equals bug", v), same, fmt.Sprintf("%d object, %d map pairs inconsistent", s.nondet["object"], s.nondet["map"]))
		}
	}
	pyviderEquals(stats[pyvider], stats[forkMain], kinds)

	w("### Examples")
	w("")
	w("Positions are attributes `k0`, `k1`, …; `U` is an unknown string. Answers are every distinct result seen in 200 comparisons; **bold** means more than one. The last row is the control: lists and tuples are visited in index order, so an unknown at a lower index always wins. That is consistent but less precise than possible, and the fix deliberately leaves it unchanged.")
	w("")
	examples := []equalsCase{
		{Kind: "object", LHS: "Uq", RHS: "pp"},
		{Kind: "map", LHS: "Uq", RHS: "pp"},
		{Kind: "object", LHS: "UUUq", RHS: "pppp"},
		{Kind: "object", LHS: "Up", RHS: "pp"},
		{Kind: "object", LHS: "qU", RHS: "pp", Marked: true},
		{Kind: "list", LHS: "Uq", RHS: "pp"},
	}
	exampleBuilds := []string{upstreamMain, forkEquals, pyvider}
	header := "| Operands | Possible concrete results | Most precise |"
	sep := "|---|---|---|"
	for _, v := range exampleBuilds {
		header += fmt.Sprintf(" `%s` |", v)
		sep += "---|"
	}
	w("%s", header)
	w("%s", sep)
	for _, ex := range examples {
		truth := equalsTruth(ex.LHS, ex.RHS)
		var ts []string
		for k := range truth {
			ts = append(ts, k)
		}
		sort.Strings(ts)
		mark := ""
		if ex.Marked {
			mark = " (left `k0` marked)"
		}
		row := fmt.Sprintf("| %s `%s` vs `%s`%s | %s | %s |", ex.Kind, ex.LHS, ex.RHS, mark, strings.Join(ts, ", "), mostPrecise(truth))
		for _, v := range exampleBuilds {
			if stats[v] == nil {
				row += " — |"
				continue
			}
			c, ok := stats[v].byKey[ex.key()]
			if !ok {
				row += " — |"
				continue
			}
			cell := categories(c)
			if len(c.Observed) > 1 {
				cell = "**" + cell + "**"
			}
			row += " " + cell + " |"
		}
		w("%s", row)
	}
	w("")

	unitTestTable(section, "equals", "TestValueEqualsUnknownAndDifference")
	tofuEqualsTables(section)
}

func resultMarked(c equalsCase) bool {
	for _, o := range c.Observed {
		if strings.Contains(o.GoString, "sensitive") {
			return true
		}
	}
	return false
}

// pyviderEquals grades pyvider-cty on the Equals cases. Its renderings are
// Python's, so it is compared with the fixed Go build by answer and by whether
// the answer carries the operand's mark, not by rendering.
func pyviderEquals(s, fixed *equalsStats, kinds []string) {
	if s == nil {
		record(pyviderSection, "Equals results present", false, "results/pyvider-cty/equals.jsonl missing — run `./run-proof.sh pyvider`")
		return
	}
	nd, un := 0, 0
	for _, k := range kinds {
		nd += s.nondet[k]
		un += s.unsound[k]
	}
	record(pyviderSection, "Equals: judge's ground truth matches the Python harness's", s.truthMismatch == 0, fmt.Sprintf("%d mismatches", s.truthMismatch))
	record(pyviderSection, "Equals: one answer for every pair", nd == 0, fmt.Sprintf("%d pairs with more than one answer", nd))
	record(pyviderSection, "Equals: never an unsound answer", un == 0, fmt.Sprintf("%d unsound", un))
	record(pyviderSection, "Equals: object/map answers are the most precise possible", s.imprecise["object"]+s.imprecise["map"] == 0, fmt.Sprintf("%d less precise", s.imprecise["object"]+s.imprecise["map"]))
	if fixed == nil {
		return
	}
	differ := 0
	var example string
	for key, fc := range fixed.byKey {
		pc, ok := s.byKey[key]
		if !ok || categories(pc) != categories(fc) || resultMarked(pc) != resultMarked(fc) {
			differ++
			if example == "" {
				example = fmt.Sprintf("— e.g. %s: fork %s, pyvider %s", key, categories(fc), categories(pc))
			}
		}
	}
	record(pyviderSection, "Equals: same answer, and same mark on the answer, as `fork-main` for every pair (lists and tuples included)", differ == 0 && len(s.byKey) == len(fixed.byKey), fmt.Sprintf("%d of %d differ %s", differ, len(fixed.byKey), example))
}

// ---------------------------------------------------------------------------
// SetProductFunc

type lengthGroup struct {
	Count  int `json:"count"`
	Length int `json:"length"`
}

type lengthCase struct {
	Groups       []lengthGroup `json:"groups"`
	Expected     string        `json:"expected"`
	FitsInt      bool          `json:"fits_int"`
	WrappedInt   int64         `json:"wrapped_int"`
	Outcome      string        `json:"outcome"`
	Length       int           `json:"length"`
	Error        string        `json:"error"`
	ResultSHA256 string        `json:"result_sha256"`
	ValuesSHA256 string        `json:"values_sha256"`
}

func (c lengthCase) spec() string {
	var parts []string
	for _, g := range c.Groups {
		parts = append(parts, fmt.Sprintf("%d×%d", g.Count, g.Length))
	}
	return strings.Join(parts, " + ")
}

func (c lengthCase) expected() *big.Int {
	n := big.NewInt(1)
	for _, g := range c.Groups {
		for range g.Count {
			n.Mul(n, big.NewInt(int64(g.Length)))
		}
	}
	return n
}

func (c lengthCase) cell() string {
	switch c.Outcome {
	case "ok":
		return fmt.Sprintf("length %d", c.Length)
	case "error":
		msg := c.Error
		if strings.HasPrefix(msg, "panic in function implementation: ") {
			return "panic: " + strings.TrimPrefix(msg, "panic in function implementation: runtime error: ")
		}
		if i := strings.Index(msg, "exceeds the safety limit of "); i >= 0 {
			return "error: over the safety limit of " + strings.TrimPrefix(msg[i:], "exceeds the safety limit of ")
		}
		return "error: " + msg
	}
	return c.Outcome
}

type diffCase struct {
	Index   int    `json:"index"`
	Call    string `json:"call"`
	Result  string `json:"result"`
	SHA256  string `json:"sha256"`
	Outcome string `json:"outcome"`
}

type memoryCase struct {
	Samples      []uint64 `json:"total_alloc_bytes"`
	Length       int      `json:"length"`
	ResultSHA256 string   `json:"result_sha256"`
	ValuesSHA256 string   `json:"values_sha256"`
}

func setProductSection() {
	const section = "SetProductFunc"
	w("## 2. `SetProductFunc` overflows its length: a wrong empty result, or a panic")
	w("")
	w("**Claim.** The result length is the product of the argument lengths, multiplied without an overflow check. At 2^64 elements it wraps to exactly zero and the function returns an empty collection; at 2^63 it wraps negative and the allocation panics. The fix returns an error when the length cannot be represented, and builds each tuple from a reused buffer instead of a second copy of the whole product.")
	w("")
	w("**Method.** Lists of N elements, K of them, with the expected length computed by the judge as an exact big integer. Separately, 20,000 randomly generated calls (lists, sets and tuples of mixed types; unknown elements, unknown and refined collections, `DynamicVal`, nulls, marks on elements and on collections) are run on every build and their full results compared byte for byte.")
	w("")

	lengths := map[string][]lengthCase{}
	for _, v := range allBuilds {
		lengths[v] = readJSONL[lengthCase](filepath.Join(resultsDir, v, "setproduct-lengths.jsonl"))
	}
	if lengths[upstreamMain] == nil || lengths[forkSetProduct] == nil {
		w("_SetProduct results missing — run `./run-proof.sh variants`._")
		record(section, "results present", false, "setproduct-lengths.jsonl missing")
		return
	}
	shown := []string{upstreamRelease, upstreamMain, forkSetProduct, forkMain, pyvider}
	header := "| Arguments (K×N) | Correct length | Fits in `int` |"
	sep := "|---|---:|---|"
	for _, v := range shown {
		header += fmt.Sprintf(" `%s` |", v)
		sep += "---|"
	}
	w("### Result length per build")
	w("")
	w("%s", header)
	w("%s", sep)
	for i, c := range lengths[upstreamMain] {
		if c.FitsInt && c.expected().Cmp(big.NewInt(100000)) <= 0 && c.Groups[0].Length != 10 && c.Groups[0].Count%4 != 0 {
			continue // most small cases are summarised by the checks below
		}
		row := fmt.Sprintf("| %s | %s | %t |", c.spec(), c.expected().String(), c.FitsInt)
		for _, v := range shown {
			if i >= len(lengths[v]) {
				row += " — |"
				continue
			}
			lc := lengths[v][i]
			cell := lc.cell()
			if lc.Outcome == "ok" && big.NewInt(int64(lc.Length)).Cmp(lc.expected()) != 0 {
				cell = "**" + cell + " (wrong)**"
			}
			row += " " + cell + " |"
		}
		w("%s", row)
	}
	w("")
	w("Rows 48×2 to 62×2 fit in an `int` but not in any computer's memory, so every build — fixed or not — fails the allocation with a recovered panic. The fix only covers lengths an `int` cannot hold; see *What this does not show*.")
	w("")
	w("`pyvider-cty` multiplies in Python integers, which cannot overflow, and refuses any product over 1,000,000 elements before building it. That limit is deliberate — its parity tracker records it as an accepted divergence from go-cty — so it answers the rows no machine can build (48×2 to 62×2) with an error instead of an allocation panic.")
	w("")
	pyviderSetProduct(lengths[pyvider], lengths[upstreamMain])

	for _, v := range []string{upstreamRelease, upstreamMain} {
		wrong, panics := 0, 0
		for _, c := range lengths[v] {
			if c.Outcome == "ok" && big.NewInt(int64(c.Length)).Cmp(c.expected()) != 0 {
				wrong++
			}
			if c.Outcome == "error" && !c.FitsInt && strings.Contains(c.Error, "panic") {
				panics++
			}
		}
		record(section, fmt.Sprintf("bug is real: `%s` returns a wrong length", v), wrong > 0, fmt.Sprintf("%d calls returned a length that is not the product", wrong))
		record(section, fmt.Sprintf("bug is real: `%s` panics on an unrepresentable length", v), panics > 0, fmt.Sprintf("%d calls", panics))
	}
	for _, v := range []string{forkSetProduct, forkMain} {
		cases := lengths[v]
		if cases == nil {
			record(section, fmt.Sprintf("`%s` results present", v), false, "missing")
			continue
		}
		wrong, badOverflow, changedSmall, compared := 0, 0, 0, 0
		for i, c := range cases {
			if c.Outcome == "ok" && big.NewInt(int64(c.Length)).Cmp(c.expected()) != 0 {
				wrong++
			}
			if !c.FitsInt && !(c.Outcome == "error" && c.Error == "result would have too many elements") && !(c.expected().Sign() == 0) {
				badOverflow++
			}
			if c.FitsInt && c.ResultSHA256 != "" {
				compared++
				if i >= len(lengths[upstreamMain]) || lengths[upstreamMain][i].ResultSHA256 != c.ResultSHA256 {
					changedSmall++
				}
			}
		}
		record(section, fmt.Sprintf("fixed: `%s` never returns a wrong length", v), wrong == 0, fmt.Sprintf("%d wrong", wrong))
		record(section, fmt.Sprintf("fixed: `%s` returns the new error for every unrepresentable length", v), badOverflow == 0, fmt.Sprintf("%d did not", badOverflow))
		record(section, fmt.Sprintf("`%s` computes every representable product identically to upstream", v), changedSmall == 0 && compared > 0, fmt.Sprintf("%d of %d results differ", changedSmall, compared))
	}
	for _, v := range []string{forkEquals, forkBounds} {
		wrong := 0
		for _, c := range lengths[v] {
			if c.Outcome == "ok" && big.NewInt(int64(c.Length)).Cmp(c.expected()) != 0 {
				wrong++
			}
		}
		if lengths[v] != nil {
			record(section, fmt.Sprintf("isolation: `%s` (no SetProduct change) still returns wrong lengths", v), wrong > 0, fmt.Sprintf("%d wrong", wrong))
		}
	}

	w("### 20,000 random calls, compared byte for byte")
	w("")
	base := readJSONL[diffCase](filepath.Join(resultsDir, upstreamMain, "setproduct-diff.jsonl"))
	w("| Build | Calls | Answered | Refused | Identical to `upstream-main` |")
	w("|---|---:|---:|---:|---:|")
	for _, v := range variants {
		cases := readJSONL[diffCase](filepath.Join(resultsDir, v, "setproduct-diff.jsonl"))
		if cases == nil {
			w("| `%s` | — | — | — | — |", v)
			continue
		}
		ok, errs, same := 0, 0, 0
		var firstDiff string
		for i, c := range cases {
			if c.Outcome == "ok" {
				ok++
			} else {
				errs++
			}
			if i < len(base) && base[i].SHA256 == c.SHA256 && base[i].Call == c.Call {
				same++
			} else if firstDiff == "" {
				firstDiff = fmt.Sprintf("call %d", c.Index)
			}
		}
		w("| `%s` | %d | %d | %d | %d |", v, len(cases), ok, errs, same)
		if v != upstreamMain {
			record(section, fmt.Sprintf("`%s` gives the same result as upstream for all 20,000 random calls", v), same == len(cases) && len(cases) == len(base), fmt.Sprintf("%d of %d identical %s", same, len(cases), firstDiff))
		}
	}
	w("| `%s` | — | — | — | not run |", pyvider)
	w("")
	w("The random calls are Go-only: they replay Go's PCG stream and compare Go renderings byte for byte, which a Python build cannot produce. pyvider-cty's own test suite compares `setproduct` with go-cty in its differential sweeps.")
	w("")

	w("### Memory allocated for 6×10 (1,000,000 tuples)")
	w("")
	w("| Build | Total allocated (5 runs) | Median | Result length | Result identical to `upstream-main` |")
	w("|---|---|---:|---:|---|")
	var baseMem memoryCase
	if m := readJSONL[memoryCase](filepath.Join(resultsDir, upstreamMain, "setproduct-memory.jsonl")); len(m) == 1 {
		baseMem = m[0]
	}
	medians := map[string]uint64{}
	for _, v := range variants {
		m := readJSONL[memoryCase](filepath.Join(resultsDir, v, "setproduct-memory.jsonl"))
		if len(m) != 1 {
			continue
		}
		samples := append([]uint64(nil), m[0].Samples...)
		sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
		medians[v] = samples[len(samples)/2]
		var mb []string
		for _, s := range m[0].Samples {
			mb = append(mb, fmt.Sprintf("%.0f", float64(s)/1e6))
		}
		w("| `%s` | %s MB | **%.0f MB** | %d | %t |", v, strings.Join(mb, ", "), float64(medians[v])/1e6, m[0].Length, m[0].ResultSHA256 == baseMem.ResultSHA256)
		if v != upstreamMain {
			record(section, fmt.Sprintf("`%s` builds an identical 1,000,000-tuple result", v), m[0].ResultSHA256 == baseMem.ResultSHA256, "sha256 "+m[0].ResultSHA256[:16])
		}
	}
	if m := readJSONL[memoryCase](filepath.Join(resultsDir, pyvider, "setproduct-memory.jsonl")); len(m) == 1 {
		same := m[0].ValuesSHA256 != "" && m[0].ValuesSHA256 == baseMem.ValuesSHA256
		w("| `%s` | not measured (Python runtime) | — | %d | %t, by element digest |", pyvider, m[0].Length, same)
		record(pyviderSection, "SetProduct: builds the same 1,000,000-tuple product as `upstream-main`, element for element", same, fmt.Sprintf("element sha256 %.16s vs %.16s", m[0].ValuesSHA256, baseMem.ValuesSHA256))
	} else {
		record(pyviderSection, "SetProduct memory-case result present", false, "missing")
	}
	w("")
	w("*Element digest*: one line per tuple, its elements joined by commas, in result order — a form both harnesses compute identically, unlike Go's `%%#v` rendering.")
	w("")
	if medians[forkSetProduct] > 0 && medians[upstreamMain] > 0 {
		record(section, "fixed: `fork-setproduct` allocates less than upstream for the same result", medians[forkSetProduct] < medians[upstreamMain], fmt.Sprintf("%.0f MB vs %.0f MB", float64(medians[forkSetProduct])/1e6, float64(medians[upstreamMain])/1e6))
	}

	unitTestTable(section, "setproduct", "TestSetproductTooManyElements")
	tofuConsoleTable(section)
}

func pyviderSetProduct(cases, upstream []lengthCase) {
	if cases == nil {
		record(pyviderSection, "SetProduct results present", false, "results/pyvider-cty/setproduct-lengths.jsonl missing — run `./run-proof.sh pyvider`")
		return
	}
	wrong, unrefused, changed, compared := 0, 0, 0, 0
	for i, c := range cases {
		if c.Outcome == "ok" && big.NewInt(int64(c.Length)).Cmp(c.expected()) != 0 {
			wrong++
		}
		if !c.FitsInt && c.Outcome != "error" {
			unrefused++
		}
		if c.Outcome == "ok" && c.ValuesSHA256 != "" && i < len(upstream) && upstream[i].ValuesSHA256 != "" {
			compared++
			if upstream[i].ValuesSHA256 != c.ValuesSHA256 {
				changed++
			}
		}
	}
	record(pyviderSection, "SetProduct: never returns a wrong length", wrong == 0, fmt.Sprintf("%d wrong", wrong))
	record(pyviderSection, "SetProduct: refuses every length an `int` cannot hold", unrefused == 0, fmt.Sprintf("%d not refused", unrefused))
	record(pyviderSection, "SetProduct: every product both compute has the same elements in the same order as `upstream-main`", changed == 0 && compared > 0, fmt.Sprintf("%d of %d differ", changed, compared))
}

// ---------------------------------------------------------------------------
// Refinement bounds

type boundsCase struct {
	Lower    string `json:"lower"`
	Upper    string `json:"upper"`
	LowerInc bool   `json:"lower_inclusive"`
	UpperInc bool   `json:"upper_inclusive"`
	Order    string `json:"order"`
	Accepted bool   `json:"accepted"`
	Panic    string `json:"panic"`
	GoString string `json:"gostring"`
}

func (c boundsCase) nonEmpty() bool {
	lo, ok1 := new(big.Rat).SetString(c.Lower)
	hi, ok2 := new(big.Rat).SetString(c.Upper)
	if !ok1 || !ok2 {
		panic("unparseable bound")
	}
	switch lo.Cmp(hi) {
	case -1:
		return true
	case 0:
		return c.LowerInc && c.UpperInc
	}
	return false
}

func (c boundsCase) text() string {
	l, u := "<", "<"
	if c.LowerInc {
		l = "≤"
	}
	if c.UpperInc {
		u = "≤"
	}
	return fmt.Sprintf("%s %s x %s %s", c.Lower, l, u, c.Upper)
}

type consequence struct {
	Probe  string `json:"probe"`
	Result string `json:"result"`
}

type consequencesCase struct {
	Accepted     bool          `json:"accepted"`
	Panic        string        `json:"panic"`
	GoString     string        `json:"gostring"`
	Consequences []consequence `json:"consequences"`
	MsgpackHex   string        `json:"msgpack_hex"`
}

type decodeCase struct {
	Description  string        `json:"description"`
	PayloadHex   string        `json:"payload_hex"`
	Outcome      string        `json:"outcome"`
	Detail       string        `json:"detail"`
	Consequences []consequence `json:"consequences"`
}

// payloadRange reads the bounds back out of a hand-encoded refinement payload
// (c7 09 0c 82 03 92 <lower> <flag> 04 92 <upper> <flag>), so the judge decides
// emptiness from the bytes rather than from the harness's description.
func payloadRange(hexPayload string) (boundsCase, bool) {
	if len(hexPayload) != 24 {
		return boundsCase{}, false
	}
	lo, err1 := strconv.ParseInt(hexPayload[12:14], 16, 64)
	hi, err2 := strconv.ParseInt(hexPayload[20:22], 16, 64)
	if err1 != nil || err2 != nil {
		return boundsCase{}, false
	}
	return boundsCase{Lower: strconv.FormatInt(lo, 10), Upper: strconv.FormatInt(hi, 10), LowerInc: hexPayload[14:16] == "c3", UpperInc: hexPayload[22:24] == "c3"}, true
}

func boundsSection() {
	const section = "Refinement bounds"
	w("## 3. Number refinement accepts `x < x` exclusive-on-both-sides, but not one side")
	w("")
	w("**Claim.** `assertConsistentBounds` picks its comparison by whether the two inclusivity flags are *equal* instead of whether both are inclusive, so `3 < x < 3` is accepted while the equally empty `3 < x ≤ 3` panics. The fix accepts equal bounds only when both are inclusive.")
	w("")
	w("**Method.** Every lower and upper bound from {−1, 0, 0.5, 1, 10^38+1}, every combination of inclusive and exclusive, applied lower-first and upper-first — 200 refinements per build. The judge decides independently, with exact rational comparison, whether each range contains any number.")
	w("")
	results := map[string][]boundsCase{}
	for _, v := range allBuilds {
		results[v] = readJSONL[boundsCase](filepath.Join(resultsDir, v, "bounds.jsonl"))
	}
	if results[upstreamMain] == nil || results[forkBounds] == nil {
		w("_Bounds results missing — run `./run-proof.sh variants`._")
		record(section, "results present", false, "bounds.jsonl missing")
		return
	}
	w("| Build | Refinements | Accepted | Accepted but empty | Refused but non-empty |")
	w("|---|---:|---:|---:|---:|")
	mismatches := map[string][]boundsCase{}
	for _, v := range allBuilds {
		cases := results[v]
		if cases == nil {
			continue
		}
		accepted, emptyAccepted, nonEmptyRefused := 0, 0, 0
		for _, c := range cases {
			if c.Accepted {
				accepted++
			}
			if c.Accepted && !c.nonEmpty() {
				emptyAccepted++
				mismatches[v] = append(mismatches[v], c)
			}
			if !c.Accepted && c.nonEmpty() {
				nonEmptyRefused++
				mismatches[v] = append(mismatches[v], c)
			}
		}
		w("| `%s` | %d | %d | %d | %d |", v, len(cases), accepted, emptyAccepted, nonEmptyRefused)
	}
	w("")

	for _, v := range []string{upstreamRelease, upstreamMain} {
		allEqualExclusive := len(mismatches[v]) > 0
		for _, c := range mismatches[v] {
			if !(c.Accepted && c.Lower == c.Upper && !c.LowerInc && !c.UpperInc) {
				allEqualExclusive = false
			}
		}
		record(section, fmt.Sprintf("bug is real: `%s` accepts empty ranges", v), len(mismatches[v]) > 0, fmt.Sprintf("%d empty ranges accepted", len(mismatches[v])))
		record(section, fmt.Sprintf("`%s` is wrong only for equal bounds exclusive on both sides", v), allEqualExclusive, fmt.Sprintf("%d mismatches, all of that shape: %t", len(mismatches[v]), allEqualExclusive))
	}
	for _, v := range []string{forkBounds, forkMain} {
		if results[v] == nil {
			record(section, fmt.Sprintf("`%s` results present", v), false, "missing")
			continue
		}
		record(section, fmt.Sprintf("fixed: `%s` accepts exactly the non-empty ranges", v), len(mismatches[v]) == 0, fmt.Sprintf("%d mismatches", len(mismatches[v])))
		changedAccepted, changedPanic := 0, 0
		for i, c := range results[v] {
			u := results[upstreamMain][i]
			if c.nonEmpty() && (c.GoString != u.GoString || c.Accepted != u.Accepted) {
				changedAccepted++
			}
			if !c.Accepted && !u.Accepted && c.Panic != u.Panic {
				changedPanic++
			}
		}
		record(section, fmt.Sprintf("`%s` builds every non-empty range identically to upstream", v), changedAccepted == 0, fmt.Sprintf("%d differ", changedAccepted))
		record(section, fmt.Sprintf("`%s` keeps upstream's panic message for every range both refuse", v), changedPanic == 0, fmt.Sprintf("%d differ", changedPanic))
	}
	for _, v := range []string{forkEquals, forkSetProduct} {
		if results[v] != nil {
			record(section, fmt.Sprintf("isolation: `%s` (no refinement change) still accepts empty ranges", v), len(mismatches[v]) > 0, fmt.Sprintf("%d", len(mismatches[v])))
		}
	}
	if py := results[pyvider]; py == nil {
		record(pyviderSection, "refinement bounds results present", false, "results/pyvider-cty/bounds.jsonl missing — run `./run-proof.sh pyvider`")
	} else {
		record(pyviderSection, "refinement builder accepts exactly the non-empty ranges", len(mismatches[pyvider]) == 0, fmt.Sprintf("%d mismatches", len(mismatches[pyvider])))
		differ := 0
		for i, c := range py {
			if i >= len(results[forkBounds]) || results[forkBounds][i].Accepted != c.Accepted {
				differ++
			}
		}
		record(pyviderSection, "refinement builder accepts and refuses exactly what `fork-bounds` does", differ == 0 && len(py) == len(results[forkBounds]), fmt.Sprintf("%d of %d differ", differ, len(py)))
	}

	same := func(a, b boundsCase) bool {
		return a.Lower == b.Lower && a.Upper == b.Upper && a.LowerInc == b.LowerInc && a.UpperInc == b.UpperInc && a.Order == b.Order
	}
	outcome := func(v string, b boundsCase) string {
		switch {
		case b.Order == "":
			return "—"
		case b.Accepted:
			return "accepted"
		case v == pyvider:
			return "raises `" + b.Panic + "`"
		}
		return "panics: `" + b.Panic + "`"
	}
	w("### The ranges upstream gets wrong")
	w("")
	w("| Range | Applied | `upstream-main` | `fork-bounds` | `pyvider-cty` |")
	w("|---|---|---|---|---|")
	for _, c := range mismatches[upstreamMain] {
		var fixed, py boundsCase
		for _, f := range results[forkBounds] {
			if same(f, c) {
				fixed = f
			}
		}
		for _, p := range results[pyvider] {
			if same(p, c) {
				py = p
			}
		}
		w("| `%s` | %s | **%s** | %s | %s |", c.text(), c.Order, outcome(upstreamMain, c), outcome(forkBounds, fixed), outcome(pyvider, py))
	}
	w("")

	w("### What the accepted empty range means")
	w("")
	w("The value upstream accepts for `3 < x < 3`, probed:")
	w("")
	w("| Build | Refinement | Probe | Result |")
	w("|---|---|---|---|")
	for _, v := range []string{upstreamMain, forkBounds, pyvider} {
		cs := readJSONL[consequencesCase](filepath.Join(resultsDir, v, "bounds-consequences.jsonl"))
		if len(cs) != 1 {
			continue
		}
		c := cs[0]
		if !c.Accepted {
			verb := "panics:"
			if v == pyvider {
				verb = "raises"
			}
			w("| `%s` | %s `%s` | — | — |", v, verb, c.Panic)
			continue
		}
		for _, p := range c.Consequences {
			w("| `%s` | accepted | `%s` | `%s` |", v, p.Probe, p.Result)
		}
		if v == upstreamMain {
			allFalse := len(c.Consequences) > 0
			for _, p := range c.Consequences {
				if strings.HasPrefix(p.Probe, "Equals(") && p.Result != "cty.False" {
					allFalse = false
				}
			}
			record(section, "bug has consequences: the accepted unknown number is unequal to every number probed, including 3", allFalse, fmt.Sprintf("%d probes", len(c.Consequences)))
			w("| `%s` | accepted | msgpack encoding | `%s` |", v, c.MsgpackHex)
		}
	}
	w("")

	w("### Decoding refinements from the wire (`msgpack.Unmarshal`)")
	w("")
	w("This is **not** part of the fix; it is recorded because the fix changes one row. go-cty's msgpack decoder applies wire bounds through the same refinement builder and does not recover its panic, so an inconsistent payload from a peer panics the decoding process. Upstream already does this for `3 < x ≤ 3` and `4 ≤ x ≤ 3`; with the fix, `3 < x < 3` behaves the same way. The payloads are hand-encoded; the `3 < x < 3` payload is byte-identical to upstream's own `msgpack.Marshal` of the value above.")
	w("")
	w("| Payload | Bytes | Contains a number | `upstream-main` | `fork-bounds` | `pyvider-cty` |")
	w("|---|---|---|---|---|---|")
	up := readJSONL[decodeCase](filepath.Join(resultsDir, upstreamMain, "bounds-decode.jsonl"))
	fx := readJSONL[decodeCase](filepath.Join(resultsDir, forkBounds, "bounds-decode.jsonl"))
	py := readJSONL[decodeCase](filepath.Join(resultsDir, pyvider, "bounds-decode.jsonl"))
	emptyDecoded, emptyPayloads, nonEmptyRefused, pyCompared := 0, 0, 0, 0
	var emptyEquals []string
	for i, c := range up {
		fixed, pyCell := "—", "—"
		rng, parsed := payloadRange(c.PayloadHex)
		nonEmpty := parsed && rng.nonEmpty()
		if parsed && !nonEmpty {
			emptyPayloads++
		}
		if i < len(fx) {
			fixed = fx[i].Outcome
		}
		if i < len(py) && py[i].PayloadHex == c.PayloadHex {
			pyCompared++
			pyCell = py[i].Outcome
			switch {
			case py[i].Outcome == "value" && parsed && !nonEmpty:
				emptyDecoded++
				pyCell = "**value (empty range kept)**"
				for _, p := range py[i].Consequences {
					if strings.HasPrefix(p.Probe, "Equals(") {
						emptyEquals = append(emptyEquals, p.Probe+" = "+p.Result)
					}
				}
			case py[i].Outcome == "error" && !nonEmpty:
				pyCell = "refused: `" + strings.Replace(py[i].Detail, "Failed to decode refined unknown payload: ", "", 1) + "`"
			case py[i].Outcome == "error":
				nonEmptyRefused++
				pyCell = "**refused (wrong)**"
			}
		}
		w("| %s | `%s` | %t | %s | %s | %s |", c.Description, c.PayloadHex, nonEmpty, c.Outcome, fixed, pyCell)
	}
	if cs := readJSONL[consequencesCase](filepath.Join(resultsDir, upstreamMain, "bounds-consequences.jsonl")); len(cs) == 1 && len(up) > 2 {
		record(section, "hand-encoded `3 < x < 3` payload equals upstream's own marshal output", cs[0].MsgpackHex == up[2].PayloadHex, cs[0].MsgpackHex)
	}
	w("")
	if len(py) > 0 {
		record(pyviderSection, "msgpack decoder refuses every payload whose range holds no number, and decodes every one that does", pyCompared == len(up) && emptyDecoded == 0 && nonEmptyRefused == 0, fmt.Sprintf("%d of %d empty ranges kept, %d non-empty refused", emptyDecoded, emptyPayloads, nonEmptyRefused))
	} else {
		record(pyviderSection, "msgpack decode results present", false, "results/pyvider-cty/bounds-decode.jsonl missing — run `./run-proof.sh pyvider`")
	}
	if len(py) > 0 && emptyDecoded == 0 && nonEmptyRefused == 0 && pyCompared == len(up) {
		meta := loadPyviderMeta()
		w("pyvider-cty at `%s` (`%.12s`) refuses all %d payloads whose range holds no number with `DeserializationError` — no panic and no empty range kept — and decodes the %d that hold one.", meta.Ref, meta.Commit, emptyPayloads, len(up)-emptyPayloads)
		w("")
	}
	if len(py) > 0 && emptyDecoded > 0 {
		sort.Strings(emptyEquals)
		w("**pyvider-cty gap found by this run — not fixed here.** pyvider-cty's decoder never crashes on these payloads, but it does not refuse them either: %d of the %d empty ranges decode to an unknown number, and those values answer %s. That is the consequence of upstream's accepted `3 < x < 3`, reached through the wire instead of the builder. pyvider-cty's refinement builder refuses all of these ranges (above); its msgpack decoder builds the refinement directly (`codec.py`, `_extract_refinements_from_payload`) and never runs that check.", emptyDecoded, emptyPayloads, "`"+strings.Join(dedupe(emptyEquals), "`, `")+"`")
		w("")
	}

	unitTestTable(section, "bounds", "TestValueRefine")
}

// ---------------------------------------------------------------------------
// Unit tests

type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

func testCounts(path, test string) (pass, fail int, failingSubtests []string) {
	seen := map[string]bool{}
	for _, e := range readJSONL[testEvent](path) {
		switch {
		case e.Test == test && e.Action == "pass":
			pass++
		case e.Test == test && e.Action == "fail":
			fail++
		case strings.HasPrefix(e.Test, test+"/") && e.Action == "fail" && !seen[e.Test]:
			seen[e.Test] = true
			failingSubtests = append(failingSubtests, strings.TrimPrefix(e.Test, test+"/"))
		}
	}
	sort.Strings(failingSubtests)
	return
}

func unitTestTable(section, name, test string) {
	w("### The fix's own regression test, before and after")
	w("")
	w("`%s`, run 20 times in one `go test -count=20`: on upstream `main` with only the fix branch's test file copied in, and on the fix branch.", test)
	w("")
	w("| Source | Passed | Failed | Failing subtests |")
	w("|---|---:|---:|---|")
	bp, bf, bs := testCounts(filepath.Join(resultsDir, "unit-tests", name+"-base.json"), test)
	fp, ff, fs := testCounts(filepath.Join(resultsDir, "unit-tests", name+"-fix.json"), test)
	w("| upstream `main` + test only | %d | %d | %s |", bp, bf, strings.Join(bs, "<br>"))
	w("| fix branch | %d | %d | %s |", fp, ff, strings.Join(fs, "<br>"))
	w("")
	record(section, fmt.Sprintf("regression test fails on unfixed code, 20 of 20 runs (`%s`)", test), bf == 20 && bp == 0, fmt.Sprintf("%d failed, %d passed", bf, bp))
	record(section, fmt.Sprintf("regression test passes on the fix, 20 of 20 runs (`%s`)", test), fp == 20 && ff == 0, fmt.Sprintf("%d passed, %d failed", fp, ff))
}

func fullSuites() {
	const section = "Upstream test suites"
	w("## Upstream's complete test suite on every fork commit")
	w("")
	w("| Fork commit (branch or `main`) | Packages passed | Packages failed | `go test` exit | `go vet` exit |")
	w("|---|---:|---:|---:|---:|")
	for _, b := range []string{"fix-equals-unknown-and-difference", "fix-setproduct-length-overflow", "fix-refinement-exclusive-equal-bounds", "main"} {
		events := readJSONL[testEvent](filepath.Join(resultsDir, "unit-tests", "full-"+b+".json"))
		pass, fail := 0, 0
		for _, e := range events {
			if e.Test == "" && e.Package != "" {
				switch e.Action {
				case "pass":
					pass++
				case "fail":
					fail++
				}
			}
		}
		exit := readText(filepath.Join(resultsDir, "unit-tests", "full-"+b+".exit"))
		vet := readText(filepath.Join(resultsDir, "unit-tests", "vet-"+b+".exit"))
		w("| `%s` | %d | %d | %s | %s |", strings.Replace(b, "-", "/", 1), pass, fail, exit, vet)
		record(section, fmt.Sprintf("all upstream tests pass on `%s`", strings.Replace(b, "-", "/", 1)), pass > 0 && fail == 0 && exit == "0", fmt.Sprintf("%d passed, %d failed, exit %s", pass, fail, exit))
		record(section, fmt.Sprintf("`go vet` clean on `%s`", strings.Replace(b, "-", "/", 1)), vet == "0", "exit "+vet)
	}
	w("")
}

// ---------------------------------------------------------------------------
// OpenTofu

func countLines(text string) map[string]int {
	ret := map[string]int{}
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			ret[l]++
		}
	}
	return ret
}

var tofuBuilds = []struct{ label, desc string }{
	{"stock", "OpenTofu v1.12.6 as installed on `PATH` (its `go.mod` pins go-cty v1.18.0)"},
	{"built-upstream-main", "OpenTofu v1.12.6 built from source on upstream go-cty `main`"},
	{"built-fork-main", "OpenTofu v1.12.6 built from source on the fork's `main`"},
}

func tofuEqualsTables(section string) {
	w("### In real OpenTofu")
	w("")
	w("`tofu/equals/main.tf` compares `{ a = <unknown until apply>, b = \"z\" }` with `{ a = \"x\", b = \"y\" }`: `b` differs, so the objects can never be equal. `tofu/count/main.tf` uses that comparison as `count = cond ? 1 : 0`. Each configuration is planned repeatedly; nothing is applied, and nothing changes between runs.")
	w("")
	w("| OpenTofu | `objects_equal` in plan | `maps_equal` in plan | `count` plan |")
	w("|---|---|---|---|")
	for _, b := range tofuBuilds {
		dir := filepath.Join(resultsDir, "tofu", b.label)
		plans := readText(filepath.Join(dir, "equals-plans.txt"))
		if plans == "" {
			w("| %s | not run | not run | not run |", b.desc)
			continue
		}
		objects, maps := map[string]int{}, map[string]int{}
		for l, n := range countLines(plans) {
			switch {
			case strings.Contains(l, "objects_equal"):
				objects[strings.TrimSpace(strings.SplitN(l, "=", 2)[1])] += n
			case strings.Contains(l, "maps_equal"):
				maps[strings.TrimSpace(strings.SplitN(l, "=", 2)[1])] += n
			}
		}
		countPlans := map[string]int{}
		for l, n := range countLines(readText(filepath.Join(dir, "count-plans.txt"))) {
			countPlans[strings.SplitN(l, ":", 2)[0]] += n
		}
		fmtCounts := func(m map[string]int) string {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var parts []string
			for _, k := range keys {
				parts = append(parts, fmt.Sprintf("%d× `%s`", m[k], k))
			}
			return strings.Join(parts, "<br>")
		}
		w("| %s | %s | %s | %s |", b.desc, fmtCounts(objects), fmtCounts(maps), fmtCounts(countPlans))
		switch b.label {
		case "stock", "built-upstream-main":
			record(section, fmt.Sprintf("bug is real in OpenTofu (%s): the same plan flips", b.label), len(objects) > 1 && countPlans["failed"] > 0 && countPlans["succeeded"] > 0, fmt.Sprintf("objects %v, count %v", objects, countPlans))
		case "built-fork-main":
			record(section, "fixed in OpenTofu built on the fork: every plan says false and succeeds", len(objects) == 1 && objects["false"] > 0 && len(maps) == 1 && maps["false"] > 0 && countPlans["failed"] == 0 && countPlans["succeeded"] > 0, fmt.Sprintf("objects %v, maps %v, count %v", objects, maps, countPlans))
		}
	}
	w("")
	if m := readText(filepath.Join(resultsDir, "tofu-built-fork-main.modules.txt")); m != "" {
		w("go-cty embedded in the from-source builds (`go version -m`):")
		w("")
		w("```")
		w("%s", readText(filepath.Join(resultsDir, "tofu-built-upstream-main.modules.txt")))
		w("%s", m)
		w("```")
		w("")
	}
}

func tofuConsoleTable(section string) {
	w("### In real OpenTofu (`tofu console`)")
	w("")
	w("| Expression | Correct | %s |", strings.Join(func() []string {
		var ret []string
		for _, b := range tofuBuilds {
			ret = append(ret, b.label)
		}
		return ret
	}(), " | "))
	w("|---|---|%s", strings.Repeat("---|", len(tofuBuilds)))
	exprs := []struct{ label, correct string }{
		{"`length(setproduct(3 × [\"a\",\"b\"]))`", "8"},
		{"`length(setproduct(8 × [256 strings]))`", "18446744073709551616"},
		{"`length(setproduct(7 × [256 strings], 1 × [128 strings]))`", "9223372036854775808"},
	}
	answers := map[string][]string{}
	for _, b := range tofuBuilds {
		text := readText(filepath.Join(resultsDir, "tofu", b.label, "console.txt"))
		var cur []string
		var got []string
		flush := func() {
			if cur != nil {
				joined := strings.TrimSpace(strings.Join(cur, " "))
				if strings.Contains(joined, "Error") {
					joined = "error: " + firstMatching(cur, "Call to function", "Error")
				}
				got = append(got, joined)
			}
		}
		for _, l := range strings.Split(text, "\n") {
			if strings.HasPrefix(l, "> ") {
				flush()
				cur = []string{}
				continue
			}
			if strings.TrimSpace(l) != "" {
				cur = append(cur, strings.TrimSpace(l))
			}
		}
		flush()
		answers[b.label] = got
	}
	for i, e := range exprs {
		row := fmt.Sprintf("| %s | %s |", e.label, e.correct)
		for _, b := range tofuBuilds {
			a := "not run"
			if i < len(answers[b.label]) {
				a = answers[b.label][i]
			}
			if a == "0" {
				a = "**0 (wrong)**"
			}
			row += " " + a + " |"
		}
		w("%s", row)
	}
	w("")
	if a := answers["stock"]; len(a) > 1 {
		record(section, "bug is real in OpenTofu (stock): setproduct of 8×256 returns 0", a[1] == "0", "got "+a[1])
	}
	if a := answers["built-fork-main"]; len(a) > 2 {
		record(section, "fixed in OpenTofu built on the fork: both overflowing products are refused, small one still 8", a[0] == "8" && strings.HasPrefix(a[1], "error") && strings.HasPrefix(a[2], "error"), strings.Join(a, " ; "))
	}
}

func dedupe(items []string) []string {
	var ret []string
	for i, s := range items {
		if i == 0 || items[i-1] != s {
			ret = append(ret, s)
		}
	}
	return ret
}

// firstMatching returns the first line containing one of the prefixes, joined
// with the line after it when that line continues the message (OpenTofu wraps
// "panic in function implementation:" onto the next line).
func firstMatching(lines []string, prefixes ...string) string {
	for _, p := range prefixes {
		for i, l := range lines {
			if strings.Contains(l, p) {
				if strings.HasSuffix(l, ":") && i+1 < len(lines) {
					return l + " " + lines[i+1]
				}
				return l
			}
		}
	}
	return strings.Join(lines, " ")
}

func limitations() {
	w("## What this does not show")
	w("")
	w("- **Equals**: operands are flat, with string members drawn from {p, q, unknown}. Nested containers, other element types and refined unknowns go through the same loops but are not enumerated here. The OpenTofu runs cover one real configuration, not Terraform itself.")
	w("- **SetProduct**: the fix adds no size limit below the range of `int`. Products that fit in `int` but not in memory (e.g. 2^40 elements) still fail the way they did before — a recovered allocation panic or an out-of-memory process — and are deliberately not run.")
	w("- **Refinement bounds**: the decoder panic on inconsistent wire payloads is pre-existing upstream behaviour, not introduced or fixed here; the fix extends it to one more payload.")
	pyMeta := loadPyviderMeta()
	w("- **pyvider-cty** is tested at `%s` (`%.12s`) as installed from GitHub, as a library only. It has no OpenTofu column, since OpenTofu evaluates expressions with go-cty, not pyvider-cty. It is not part of the random-call replay or the allocation measurements, which are specific to Go. Its own test suite is not run here.", pyMeta.Ref, pyMeta.Commit)
	w("- The upstream issues are not filed and these branches are not proposed upstream.")
	w("")
}
