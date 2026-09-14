# go-cty correctness fixes — reproducible proof

Evidence that three defects in [go-cty](https://github.com/zclconf/go-cty) exist
in both the v1.19.0 release and current upstream `main`, and that the fixes on
the [`livingstaccato/go-cty`](https://github.com/livingstaccato/go-cty) fork
remove them without changing anything else. The same cases are also run against
[pyvider-cty](https://github.com/provide-io/pyvider-cty), the Python port of
go-cty.

**The verdict is [`REPORT.md`](REPORT.md).** Every number and PASS/FAIL in it is
computed from the raw observations committed in [`results/`](results/), and
`./run-proof.sh all` regenerates both from nothing.

## The three defects

| Defect | Fork PR | Fix tested at |
| --- | --- | --- |
| `Value.Equals` answers `false` or unknown at random for an object or map holding both an unknown and a definite difference | [#1](https://github.com/livingstaccato/go-cty/pull/1) | `ef2c1cac` |
| A number refinement accepts an empty range when both bounds are equal and exclusive, and panics when only one is | [#2](https://github.com/livingstaccato/go-cty/pull/2) | `781f24c6` |
| `SetProductFunc`'s result length overflows, returning an empty result or panicking | [#3](https://github.com/livingstaccato/go-cty/pull/3) | `e7c9522b` |

Each is also filed upstream, since pull requests are disabled on
[`zclconf/go-cty`](https://github.com/zclconf/go-cty): issues
[#226](https://github.com/zclconf/go-cty/issues/226),
[#227](https://github.com/zclconf/go-cty/issues/227) and
[#228](https://github.com/zclconf/go-cty/issues/228), each linking the merged
fork PR and this repo.

## What is built and checked

| Build | Source |
| --- | --- |
| `upstream-v1.19.0` | go-cty v1.19.0, the release |
| `upstream-main` | go-cty `main` at `a918e117` |
| `fork-equals`, `fork-setproduct`, `fork-bounds` | each fix on its own |
| `fork-main` | the fork's `main` at `d8ae3d7f`, where all three pull requests merged |
| `pyvider-cty` | pyvider-cty at tag `v0.6.1` |
| OpenTofu v1.12.6 | stock, and built from source twice: on upstream `main` and on the fork's `main` |

One Go program (`proof/harness`) is built against every go-cty variant and only
records what go-cty answers. A Python program (`proof/pyharness`) records the
same cases for pyvider-cty. A separate judge (`proof/judge`, standard library
only, no go-cty import) works out the expected answers independently and grades
every build. It checks that:

- each bug is real on both upstream builds, and the `Equals` and `setproduct`
  bugs also show in stock OpenTofu;
- each fix removes its bug, and the combined build removes all three;
- each single-fix build still has the other two bugs, so every fix is measured
  on its own;
- nothing else changes: identical results for every other input, including
  20,000 random `setproduct` calls and a 1,000,000-tuple product;
- each fix's regression test fails 20 of 20 runs on unfixed code and passes
  20 of 20 on the fix, and go-cty's whole test suite and `go vet` pass on every
  fork commit;
- each pull request's head is the commit that was tested, or only merges `main`
  into it;
- pyvider-cty gives the fixed answers, and was installed at the commit GitHub has
  for its tag.

## Why the fork builds are pinned to those commits

Every fork build is pinned to a commit on the fork's `main`, so it stays
reachable after branches are deleted. Pull requests #2 and #3 each merged `main`
into their branch before merging, so their final heads already carry the fixes
merged before them. A single-fix build pinned there would not be a single fix,
so each is pinned to that head's first parent, the commit that carries its fix
alone. The report checks that relationship against GitHub.

## Requirements

- macOS or Linux, with bash, git, `tar` and `awk`
- Go 1.25 or newer (the committed results used go1.27.1). OpenTofu's build
  fetches the exact Go toolchain its `go.mod` names.
- [OpenTofu](https://opentofu.org) **1.12.6** on `PATH`, for the stock runs
- [uv](https://docs.astral.sh/uv/), which fetches Python 3.13 itself
- [jq](https://jqlang.org)
- [gh](https://cli.github.com), logged in (`gh auth login`) or given a token in
  `GH_TOKEN`, to read the pull requests
- Network access to GitHub, `proxy.golang.org`, `sum.golang.org` and PyPI
- About 1 GB free for `build/`

`run-proof.sh` checks for these before it starts and names anything missing.

## Running it

```sh
./run-proof.sh all
```

Run it from anywhere; it works relative to its own directory. `all` clears
`results/` first. It takes about six minutes with warm Go and uv caches; a first
run on a machine also downloads go-cty, OpenTofu's source and dependencies, a Go
toolchain and Python, so allow longer. Each step can also run on its own:

| Step | What it does |
| --- | --- |
| `prs` | reads the fork's pull requests and their heads' parents from GitHub |
| `variants` | builds and runs the Go harness against every go-cty build |
| `pyvider` | installs pyvider-cty and runs the Python harness |
| `unit-tests` | the three regression tests, then go-cty's full suite and `go vet` |
| `tofu-stock` | plans the OpenTofu configurations in `tofu/` with the installed `tofu` |
| `tofu-built` | builds OpenTofu twice from source and plans the same configurations |
| `report` | regenerates `REPORT.md` from `results/`; needs only Go |

Optional environment variables: `PLAN_RUNS` (OpenTofu plans per configuration,
default 30), `PYHARNESS_EQUALS_ITERATIONS` (default 200), `PYVIDER_REF` (a
pyvider-cty tag or branch, default `v0.6.1`), and `FORK_CLONE` (an existing clone
of the fork to use instead of `build/go-cty`).

## On GitHub Actions

`.github/workflows/proof.yml` runs `./run-proof.sh all` on a fresh Ubuntu runner
with the same pinned Go, OpenTofu and uv. Start it from the Actions tab with **Run
workflow**. The job fails if any check fails, puts `REPORT.md` on the run's
summary page, and attaches `results/` and `REPORT.md` as the `proof` artifact.
It does not commit anything back; the committed `results/` stay the ones
`REPORT.md` was generated from.

## What differs from run to run

The raw files are observations, not fixtures. Between two runs of the same
pinned code, the Go harness and pyvider-cty outputs are byte-identical except
`setproduct-memory.jsonl`, which holds allocation measurements. The OpenTofu
results differ: how many of the plans land on each of unfixed go-cty's two
`Equals` answers changes from run to run, because the bug follows Go's
randomised map iteration order, and the panic traces in `console.txt` carry
memory addresses. The `go test` JSON carries timestamps. The checks are written
to hold regardless ("the answer varies", not "it varied 14 times"), and every
build's module version and `go.sum` hash is recorded in `REPORT.md`, so a rerun
can confirm it built the same code.

The raw output also contains absolute paths from the machine that produced it:
the Go module cache in `module.json` and in panic traces, and the source tree in
test output. No check reads them.

## Layout

| Path | What it is |
| --- | --- |
| `REPORT.md` | The verdict, generated by `proof/judge` |
| `results/` | Raw observations from the run that produced `REPORT.md` |
| `run-proof.sh` | Every step, and every pinned commit and source URL |
| `proof/harness/main.go` | The only program that calls go-cty |
| `proof/pyharness/harness.py` | The same cases against pyvider-cty |
| `proof/judge/main.go` | Standard library only: recomputes the expected answers and grades every build |
| `tofu/equals`, `tofu/count` | OpenTofu configurations, planned repeatedly and never applied |
| `build/`, `work/` | Not committed: builds, clones and scratch source trees, all regenerated |
