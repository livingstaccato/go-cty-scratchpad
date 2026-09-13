# go-cty correctness fixes — proof report

Generated 2026-09-13 11:42 PDT by `run-proof.sh`. Every number, outcome and PASS/FAIL below is computed from the files in `results/`.

Three defects in go-cty, each shown to exist in the released and current upstream code, and shown to be fixed on the `livingstaccato/go-cty` fork. The same harness source is built against every variant; a separate judge (standard library only, no go-cty import) computes the expected answers independently and grades every build.

Alongside the Go builds, **pyvider-cty** — the Python port of go-cty — is put through the same cases by a Python harness and graded by the same judge, so every section also shows where the Python implementation stands.

## Verdict

| Area | Checks | Passed | Result |
|---|---:|---:|---|
| Fork branches | 3 | 3 | ✅ PASS |
| Value.Equals | 29 | 29 | ✅ PASS |
| SetProductFunc | 27 | 27 | ✅ PASS |
| Refinement bounds | 16 | 16 | ✅ PASS |
| Upstream test suites | 8 | 8 | ✅ PASS |
| pyvider-cty | 13 | 13 | ✅ PASS |

<details><summary>Every check</summary>

| Area | Check | Result | Evidence |
|---|---|---|---|
| pyvider-cty | installed package is the commit GitHub had at run time | ✅ PASS | GitHub 858fef516aa9, installed 858fef516aa9 |
| Fork branches | PR #1 head is the tested commit, or only merges its base into it | ✅ PASS | PR ef2c1cac99f5, tested ef2c1cac99f5, head |
| Fork branches | PR #2 head is the tested commit, or only merges its base into it | ✅ PASS | PR e4591f58f99b, tested 781f24c6bf5e, head merges base into it |
| Fork branches | PR #3 head is the tested commit, or only merges its base into it | ✅ PASS | PR f2d0e3507ffc, tested e7c9522b1c20, head merges base into it |
| Value.Equals | judge's ground truth matches harness's (`upstream-v1.19.0`) | ✅ PASS | 0 mismatches |
| Value.Equals | judge's ground truth matches harness's (`upstream-main`) | ✅ PASS | 0 mismatches |
| Value.Equals | judge's ground truth matches harness's (`fork-equals`) | ✅ PASS | 0 mismatches |
| Value.Equals | judge's ground truth matches harness's (`fork-setproduct`) | ✅ PASS | 0 mismatches |
| Value.Equals | judge's ground truth matches harness's (`fork-bounds`) | ✅ PASS | 0 mismatches |
| Value.Equals | judge's ground truth matches harness's (`fork-main`) | ✅ PASS | 0 mismatches |
| Value.Equals | bug is real: `upstream-v1.19.0` answers objects inconsistently | ✅ PASS | 4620 object pairs, 4620 map pairs gave two different answers |
| Value.Equals | bug is real: `upstream-v1.19.0` answers maps inconsistently | ✅ PASS | 4620 map pairs |
| Value.Equals | lists and tuples were already consistent on `upstream-v1.19.0` | ✅ PASS | 0 list, 0 tuple |
| Value.Equals | bug is real: `upstream-main` answers objects inconsistently | ✅ PASS | 4620 object pairs, 4620 map pairs gave two different answers |
| Value.Equals | bug is real: `upstream-main` answers maps inconsistently | ✅ PASS | 4620 map pairs |
| Value.Equals | lists and tuples were already consistent on `upstream-main` | ✅ PASS | 0 list, 0 tuple |
| Value.Equals | fixed: `fork-equals` gives one answer for every pair | ✅ PASS | 0 pairs with more than one answer |
| Value.Equals | fixed: `fork-equals` never gives an unsound answer | ✅ PASS | 0 unsound |
| Value.Equals | fixed: `fork-equals` object/map answers are the most precise possible | ✅ PASS | 0 less precise |
| Value.Equals | `fork-equals` never answers anything upstream never answered (marks included) | ✅ PASS | 0 new answers  |
| Value.Equals | `fork-equals` leaves every list and tuple answer exactly as upstream | ✅ PASS | 0 changed |
| Value.Equals | fixed: `fork-main` gives one answer for every pair | ✅ PASS | 0 pairs with more than one answer |
| Value.Equals | fixed: `fork-main` never gives an unsound answer | ✅ PASS | 0 unsound |
| Value.Equals | fixed: `fork-main` object/map answers are the most precise possible | ✅ PASS | 0 less precise |
| Value.Equals | `fork-main` never answers anything upstream never answered (marks included) | ✅ PASS | 0 new answers  |
| Value.Equals | `fork-main` leaves every list and tuple answer exactly as upstream | ✅ PASS | 0 changed |
| Value.Equals | isolation: `fork-setproduct` (no Equals change) still shows the Equals bug | ✅ PASS | 4620 object, 4620 map pairs inconsistent |
| Value.Equals | isolation: `fork-bounds` (no Equals change) still shows the Equals bug | ✅ PASS | 4620 object, 4620 map pairs inconsistent |
| pyvider-cty | Equals: judge's ground truth matches the Python harness's | ✅ PASS | 0 mismatches |
| pyvider-cty | Equals: one answer for every pair | ✅ PASS | 0 pairs with more than one answer |
| pyvider-cty | Equals: never an unsound answer | ✅ PASS | 0 unsound |
| pyvider-cty | Equals: object/map answers are the most precise possible | ✅ PASS | 0 less precise |
| pyvider-cty | Equals: same answer, and same mark on the answer, as `fork-main` for every pair (lists and tuples included) | ✅ PASS | 0 of 32796 differ  |
| Value.Equals | regression test fails on unfixed code, 20 of 20 runs (`TestValueEqualsUnknownAndDifference`) | ✅ PASS | 20 failed, 0 passed |
| Value.Equals | regression test passes on the fix, 20 of 20 runs (`TestValueEqualsUnknownAndDifference`) | ✅ PASS | 20 passed, 0 failed |
| Value.Equals | bug is real in OpenTofu (stock): the same plan flips | ✅ PASS | objects map[(known after apply):15 false:15], count map[failed:23 succeeded:7] |
| Value.Equals | bug is real in OpenTofu (built-upstream-main): the same plan flips | ✅ PASS | objects map[(known after apply):20 false:10], count map[failed:21 succeeded:9] |
| Value.Equals | fixed in OpenTofu built on the fork: every plan says false and succeeds | ✅ PASS | objects map[false:30], maps map[false:30], count map[succeeded:30] |
| pyvider-cty | SetProduct: never returns a wrong length | ✅ PASS | 0 wrong |
| pyvider-cty | SetProduct: refuses every length an `int` cannot hold | ✅ PASS | 0 not refused |
| pyvider-cty | SetProduct: every product both compute has the same elements in the same order as `upstream-main` | ✅ PASS | 0 of 25 differ |
| SetProductFunc | bug is real: `upstream-v1.19.0` returns a wrong length | ✅ PASS | 6 calls returned a length that is not the product |
| SetProductFunc | bug is real: `upstream-v1.19.0` panics on an unrepresentable length | ✅ PASS | 4 calls |
| SetProductFunc | bug is real: `upstream-main` returns a wrong length | ✅ PASS | 6 calls returned a length that is not the product |
| SetProductFunc | bug is real: `upstream-main` panics on an unrepresentable length | ✅ PASS | 4 calls |
| SetProductFunc | fixed: `fork-setproduct` never returns a wrong length | ✅ PASS | 0 wrong |
| SetProductFunc | fixed: `fork-setproduct` returns the new error for every unrepresentable length | ✅ PASS | 0 did not |
| SetProductFunc | `fork-setproduct` computes every representable product identically to upstream | ✅ PASS | 0 of 25 results differ |
| SetProductFunc | fixed: `fork-main` never returns a wrong length | ✅ PASS | 0 wrong |
| SetProductFunc | fixed: `fork-main` returns the new error for every unrepresentable length | ✅ PASS | 0 did not |
| SetProductFunc | `fork-main` computes every representable product identically to upstream | ✅ PASS | 0 of 25 results differ |
| SetProductFunc | isolation: `fork-equals` (no SetProduct change) still returns wrong lengths | ✅ PASS | 6 wrong |
| SetProductFunc | isolation: `fork-bounds` (no SetProduct change) still returns wrong lengths | ✅ PASS | 6 wrong |
| SetProductFunc | `upstream-v1.19.0` gives the same result as upstream for all 20,000 random calls | ✅ PASS | 20000 of 20000 identical  |
| SetProductFunc | `fork-equals` gives the same result as upstream for all 20,000 random calls | ✅ PASS | 20000 of 20000 identical  |
| SetProductFunc | `fork-setproduct` gives the same result as upstream for all 20,000 random calls | ✅ PASS | 20000 of 20000 identical  |
| SetProductFunc | `fork-bounds` gives the same result as upstream for all 20,000 random calls | ✅ PASS | 20000 of 20000 identical  |
| SetProductFunc | `fork-main` gives the same result as upstream for all 20,000 random calls | ✅ PASS | 20000 of 20000 identical  |
| SetProductFunc | `upstream-v1.19.0` builds an identical 1,000,000-tuple result | ✅ PASS | sha256 5fcd3ea48168f967 |
| SetProductFunc | `fork-equals` builds an identical 1,000,000-tuple result | ✅ PASS | sha256 5fcd3ea48168f967 |
| SetProductFunc | `fork-setproduct` builds an identical 1,000,000-tuple result | ✅ PASS | sha256 5fcd3ea48168f967 |
| SetProductFunc | `fork-bounds` builds an identical 1,000,000-tuple result | ✅ PASS | sha256 5fcd3ea48168f967 |
| SetProductFunc | `fork-main` builds an identical 1,000,000-tuple result | ✅ PASS | sha256 5fcd3ea48168f967 |
| pyvider-cty | SetProduct: builds the same 1,000,000-tuple product as `upstream-main`, element for element | ✅ PASS | element sha256 c757c8dbcf24c0e7 vs c757c8dbcf24c0e7 |
| SetProductFunc | fixed: `fork-setproduct` allocates less than upstream for the same result | ✅ PASS | 288 MB vs 504 MB |
| SetProductFunc | regression test fails on unfixed code, 20 of 20 runs (`TestSetproductTooManyElements`) | ✅ PASS | 20 failed, 0 passed |
| SetProductFunc | regression test passes on the fix, 20 of 20 runs (`TestSetproductTooManyElements`) | ✅ PASS | 20 passed, 0 failed |
| SetProductFunc | bug is real in OpenTofu (stock): setproduct of 8×256 returns 0 | ✅ PASS | got 0 |
| SetProductFunc | fixed in OpenTofu built on the fork: both overflowing products are refused, small one still 8 | ✅ PASS | 8 ; error: Call to function "setproduct" failed: result would have too many elements. ; error: Call to function "setproduct" failed: result would have too many elements. |
| Refinement bounds | bug is real: `upstream-v1.19.0` accepts empty ranges | ✅ PASS | 10 empty ranges accepted |
| Refinement bounds | `upstream-v1.19.0` is wrong only for equal bounds exclusive on both sides | ✅ PASS | 10 mismatches, all of that shape: true |
| Refinement bounds | bug is real: `upstream-main` accepts empty ranges | ✅ PASS | 10 empty ranges accepted |
| Refinement bounds | `upstream-main` is wrong only for equal bounds exclusive on both sides | ✅ PASS | 10 mismatches, all of that shape: true |
| Refinement bounds | fixed: `fork-bounds` accepts exactly the non-empty ranges | ✅ PASS | 0 mismatches |
| Refinement bounds | `fork-bounds` builds every non-empty range identically to upstream | ✅ PASS | 0 differ |
| Refinement bounds | `fork-bounds` keeps upstream's panic message for every range both refuse | ✅ PASS | 0 differ |
| Refinement bounds | fixed: `fork-main` accepts exactly the non-empty ranges | ✅ PASS | 0 mismatches |
| Refinement bounds | `fork-main` builds every non-empty range identically to upstream | ✅ PASS | 0 differ |
| Refinement bounds | `fork-main` keeps upstream's panic message for every range both refuse | ✅ PASS | 0 differ |
| Refinement bounds | isolation: `fork-equals` (no refinement change) still accepts empty ranges | ✅ PASS | 10 |
| Refinement bounds | isolation: `fork-setproduct` (no refinement change) still accepts empty ranges | ✅ PASS | 10 |
| pyvider-cty | refinement builder accepts exactly the non-empty ranges | ✅ PASS | 0 mismatches |
| pyvider-cty | refinement builder accepts and refuses exactly what `fork-bounds` does | ✅ PASS | 0 of 200 differ |
| Refinement bounds | bug has consequences: the accepted unknown number is unequal to every number probed, including 3 | ✅ PASS | 6 probes |
| Refinement bounds | hand-encoded `3 < x < 3` payload equals upstream's own marshal output | ✅ PASS | c7090c82039203c2049203c2 |
| pyvider-cty | msgpack decoder refuses every payload whose range holds no number, and decodes every one that does | ✅ PASS | 0 of 3 empty ranges kept, 0 non-empty refused |
| Refinement bounds | regression test fails on unfixed code, 20 of 20 runs (`TestValueRefine`) | ✅ PASS | 20 failed, 0 passed |
| Refinement bounds | regression test passes on the fix, 20 of 20 runs (`TestValueRefine`) | ✅ PASS | 20 passed, 0 failed |
| Upstream test suites | all upstream tests pass on `fix/equals-unknown-and-difference` | ✅ PASS | 9 passed, 0 failed, exit 0 |
| Upstream test suites | `go vet` clean on `fix/equals-unknown-and-difference` | ✅ PASS | exit 0 |
| Upstream test suites | all upstream tests pass on `fix/setproduct-length-overflow` | ✅ PASS | 9 passed, 0 failed, exit 0 |
| Upstream test suites | `go vet` clean on `fix/setproduct-length-overflow` | ✅ PASS | exit 0 |
| Upstream test suites | all upstream tests pass on `fix/refinement-exclusive-equal-bounds` | ✅ PASS | 9 passed, 0 failed, exit 0 |
| Upstream test suites | `go vet` clean on `fix/refinement-exclusive-equal-bounds` | ✅ PASS | exit 0 |
| Upstream test suites | all upstream tests pass on `main` | ✅ PASS | 9 passed, 0 failed, exit 0 |
| Upstream test suites | `go vet` clean on `main` | ✅ PASS | exit 0 |

</details>

## Reproducing this report

```sh
./run-proof.sh all        # from this directory; or one step: prs | variants | pyvider | unit-tests | tofu-stock | tofu-built | report
```

Needs git, Go 1.25 or newer, OpenTofu 1.12.6 on `PATH`, uv, jq, and gh logged in (`gh auth login`); `README.md` has the details. Everything else is fetched over https into `build/`.

| Path | What it is |
|---|---|
| `proof/harness/main.go` | The only program that calls go-cty. Built once per variant; prints observations, contains no expectations. |
| `proof/pyharness/harness.py` | The same cases against pyvider-cty, printed in the same shapes. Contains no expectations. |
| `proof/judge/main.go` | Standard library only. Recomputes ground truth, grades each build, writes this file. |
| `tofu/equals`, `tofu/count` | OpenTofu configurations, planned repeatedly (never applied). |
| `results/` | Raw observations from the last run. |
| `build/` | Not committed. Every go-cty build, the fork's clone (`build/go-cty`), OpenTofu v1.12.6 source (`build/opentofu`) and the two OpenTofu builds from it. |
| `work/` | Not committed. The source trees the regression tests and full suites run in. |

## Provenance

Each build fetched go-cty through `proxy.golang.org`; the `h1:` hashes are the ones `go mod verify` checked, and for public modules they are also recorded in `sum.golang.org`. Fork branches are resolved to the commit the fork's remote had at run time, so a result cannot come from an unpushed local edit.

| Variant | Source | Branch / ref | Commit | Module actually built | `go.sum` hash | Go |
|---|---|---|---|---|---|---|
| `upstream-v1.19.0` | upstream | `v1.19.0` | `v1.19.0` | `github.com/zclconf/go-cty@v1.19.0` | `h1:IV8WdqYZc2c5rLX9bEoLNXKojBAp0MZPBHMIrCoa/s4=` | go1.27.1 |
| `upstream-main` | upstream | `a918e1174fcf2a25b7a222e7e78b00ea40ace26c` | `a918e1174fcf` | `github.com/zclconf/go-cty@v1.19.1-0.20260821180702-a918e1174fcf` | `h1:w3R6RP4NipXHPQjI09QaR/AmBAON0lHNDGcHmay4rJQ=` | go1.27.1 |
| `fork-equals` | fork | `fix/equals-unknown-and-difference` | `ef2c1cac99f5` | `github.com/livingstaccato/go-cty@v0.0.0-20260913001827-ef2c1cac99f5` | `h1:TztGooEpm21caZomxhWiVxbpmFu/UooONlnlnRCAqis=` | go1.27.1 |
| `fork-setproduct` | fork | `fix/setproduct-length-overflow` | `e7c9522b1c20` | `github.com/livingstaccato/go-cty@v0.0.0-20260913001832-e7c9522b1c20` | `h1:E8/N7MXuQMVzfUKidQbOAzwdiskvoHKmNjgwFYOxVO8=` | go1.27.1 |
| `fork-bounds` | fork | `fix/refinement-exclusive-equal-bounds` | `781f24c6bf5e` | `github.com/livingstaccato/go-cty@v0.0.0-20260913001857-781f24c6bf5e` | `h1:ZaXDdMfKd+vVaK7HqYvk0q5Bm5j3i4XKJqOv2JeI/MM=` | go1.27.1 |
| `fork-main` | fork | `main` | `d8ae3d7fbbe2` | `github.com/livingstaccato/go-cty@v0.0.0-20260913020742-d8ae3d7fbbe2` | `h1:P323qMe9oveIcAl8PXWNsYWHavSKF0+KNHuaqMDuZJY=` | go1.27.1 |

pyvider-cty is installed with `uv pip install` straight from GitHub, at the commit its branch had at run time; the installed commit is read back from the package's own `direct_url.json`.

| Build | Branch | Commit on GitHub | Commit installed | Package | Python |
|---|---|---|---|---|---|
| `pyvider-cty` | `v0.6.1` | `858fef516aa9` | `858fef516aa9` | `pyvider-cty 0.6.1` | 3.13.3 |

## The fork's pull requests carry exactly what was tested

Pull requests as GitHub reports them (`gh pr list -R livingstaccato/go-cty`), against the commit each build above was pinned to for the same branch. A PR head that only merged its base branch into the tested commit also counts: its first parent is the tested commit and its second is already on the base. Each fix is tested at that first parent, so it is measured on its own, without the fixes merged before it.

| PR | Title | Branch → base | PR head | Commit tested | Relation | Result |
|---|---|---|---|---|---|---|
| [#1](https://github.com/livingstaccato/go-cty/pull/1) | `Value.Equals` is non-deterministic for objects and maps holding both an unknown and a definite difference | `fix/equals-unknown-and-difference` → `main` | `ef2c1cac99f5` | `ef2c1cac99f5` | head | ✅ PASS |
| [#2](https://github.com/livingstaccato/go-cty/pull/2) | Number refinement accepts an empty range when both bounds are equal and exclusive, but panics when only one is | `fix/refinement-exclusive-equal-bounds` → `main` | `e4591f58f99b` | `781f24c6bf5e` | head merges base into it | ✅ PASS |
| [#3](https://github.com/livingstaccato/go-cty/pull/3) | stdlib: `SetProductFunc` result length overflows, returning an empty result or panicking | `fix/setproduct-length-overflow` → `main` | `f2d0e3507ffc` | `e7c9522b1c20` | head merges base into it | ✅ PASS |
| — | all three pull requests merged | `main` | — | `d8ae3d7fbbe2` | (no PR) | — |

## 1. `Value.Equals` gives different answers for the same objects and maps

**Claim.** When an object or map holds an unknown member *and* a member that definitely differs, upstream `Equals` answers `false` or unknown depending on Go's randomized map iteration order. The fix makes a definite difference win, so the answer is always `false`.

**Method.** Every pair of operands with 1–4 positions, each position one of `p`, `q` or unknown (`U`), for objects, maps, lists and tuples — 29,520 unmarked pairs, plus 3,276 with a `sensitive` mark on the left operand's first position. Each pair is compared **200 times** in one process. Ground truth is computed by the judge, not by go-cty: every unknown is replaced by each of `p`, `q` and a fresh value `r` in every combination, and the set of possible concrete results is collected. An answer is *sound* if no substitution contradicts it; the *most precise* answer is the concrete result when all substitutions agree, and unknown otherwise.

### Results per build

| Build | Pairs | Different answers on repeat (object / map / list / tuple) | Unsound answers | Object/map answers less precise than possible |
|---|---:|---|---:|---:|
| `upstream-v1.19.0` | 32796 | 4620 / 4620 / 0 / 0 | 0 | 9240 |
| `upstream-main` | 32796 | 4620 / 4620 / 0 / 0 | 0 | 9240 |
| `fork-equals` | 32796 | 0 / 0 / 0 / 0 | 0 | 0 |
| `fork-setproduct` | 32796 | 4620 / 4620 / 0 / 0 | 0 | 9240 |
| `fork-bounds` | 32796 | 4620 / 4620 / 0 / 0 | 0 | 9240 |
| `fork-main` | 32796 | 0 / 0 / 0 / 0 | 0 | 0 |
| `pyvider-cty` | 32796 | 0 / 0 / 0 / 0 | 0 | 0 |

### Examples

Positions are attributes `k0`, `k1`, …; `U` is an unknown string. Answers are every distinct result seen in 200 comparisons; **bold** means more than one. The last row is the control: lists and tuples are visited in index order, so an unknown at a lower index always wins. That is consistent but less precise than possible, and the fix deliberately leaves it unchanged.

| Operands | Possible concrete results | Most precise | `upstream-main` | `fork-equals` | `pyvider-cty` |
|---|---|---|---|---|---|
| object `Uq` vs `pp` | false | false | **false / unknown** | false | false |
| map `Uq` vs `pp` | false | false | **false / unknown** | false | false |
| object `UUUq` vs `pppp` | false | false | **false / unknown** | false | false |
| object `Up` vs `pp` | false, true | unknown | unknown | unknown | unknown |
| object `qU` vs `pp` (left `k0` marked) | false | false | **false / unknown** | false | false |
| list `Uq` vs `pp` | false | false | unknown | unknown | unknown |

### The fix's own regression test, before and after

`TestValueEqualsUnknownAndDifference`, run 20 times in one `go test -count=20`: on upstream `main` with only the fix branch's test file copied in, and on the fix branch.

| Source | Passed | Failed | Failing subtests |
|---|---:|---:|---|
| upstream `main` + test only | 0 | 20 | map_with_an_unknown_element_and_a_different_element<br>object_with_an_unknown_attribute_and_a_different_attribute |
| fix branch | 20 | 0 |  |

### In real OpenTofu

`tofu/equals/main.tf` compares `{ a = <unknown until apply>, b = "z" }` with `{ a = "x", b = "y" }`: `b` differs, so the objects can never be equal. `tofu/count/main.tf` uses that comparison as `count = cond ? 1 : 0`. Each configuration is planned repeatedly; nothing is applied, and nothing changes between runs.

| OpenTofu | `objects_equal` in plan | `maps_equal` in plan | `count` plan |
|---|---|---|---|
| OpenTofu v1.12.6 as installed on `PATH` (its `go.mod` pins go-cty v1.18.0) | 15× `(known after apply)`<br>15× `false` | 22× `(known after apply)`<br>8× `false` | 23× `failed`<br>7× `succeeded` |
| OpenTofu v1.12.6 built from source on upstream go-cty `main` | 20× `(known after apply)`<br>10× `false` | 26× `(known after apply)`<br>4× `false` | 21× `failed`<br>9× `succeeded` |
| OpenTofu v1.12.6 built from source on the fork's `main` | 30× `false` | 30× `false` | 30× `succeeded` |

go-cty embedded in the from-source builds (`go version -m`):

```
/Volumes/data/pyv/go-cty-scratchpad/build/tofu-upstream-main/tofu: go1.26.6
	dep	github.com/zclconf/go-cty	v1.19.1-0.20260821180702-a918e1174fcf	h1:w3R6RP4NipXHPQjI09QaR/AmBAON0lHNDGcHmay4rJQ=
	dep	github.com/zclconf/go-cty-debug	v0.0.0-20240509010212-0d6042c53940	h1:4r45xpDWB6ZMSMNJFMOjqrGHynW3DIBuR2H9j0ug+Mo=
	dep	github.com/zclconf/go-cty-yaml	v1.2.0	h1:GDyL4+e/Qe/S0B7YaecMLbVvAR/Mp21CXMOSiCTOi1M=
/Volumes/data/pyv/go-cty-scratchpad/build/tofu-fork-main/tofu: go1.26.6
	dep	github.com/zclconf/go-cty	v1.18.0
	=>	github.com/livingstaccato/go-cty	v0.0.0-20260913020742-d8ae3d7fbbe2	h1:P323qMe9oveIcAl8PXWNsYWHavSKF0+KNHuaqMDuZJY=
	dep	github.com/zclconf/go-cty-debug	v0.0.0-20240509010212-0d6042c53940	h1:4r45xpDWB6ZMSMNJFMOjqrGHynW3DIBuR2H9j0ug+Mo=
	dep	github.com/zclconf/go-cty-yaml	v1.2.0	h1:GDyL4+e/Qe/S0B7YaecMLbVvAR/Mp21CXMOSiCTOi1M=
```

## 2. `SetProductFunc` overflows its length: a wrong empty result, or a panic

**Claim.** The result length is the product of the argument lengths, multiplied without an overflow check. At 2^64 elements it wraps to exactly zero and the function returns an empty collection; at 2^63 it wraps negative and the allocation panics. The fix returns an error when the length cannot be represented, and builds each tuple from a reused buffer instead of a second copy of the whole product.

**Method.** Lists of N elements, K of them, with the expected length computed by the judge as an exact big integer. Separately, 20,000 randomly generated calls (lists, sets and tuples of mixed types; unknown elements, unknown and refined collections, `DynamicVal`, nulls, marks on elements and on collections) are run on every build and their full results compared byte for byte.

### Result length per build

| Arguments (K×N) | Correct length | Fits in `int` | `upstream-v1.19.0` | `upstream-main` | `fork-setproduct` | `fork-main` | `pyvider-cty` |
|---|---:|---|---|---|---|---|---|
| 4×2 | 16 | true | length 16 | length 16 | length 16 | length 16 | length 16 |
| 8×2 | 256 | true | length 256 | length 256 | length 256 | length 256 | length 256 |
| 12×2 | 4096 | true | length 4096 | length 4096 | length 4096 | length 4096 | length 4096 |
| 16×2 | 65536 | true | length 65536 | length 65536 | length 65536 | length 65536 | length 65536 |
| 4×3 | 81 | true | length 81 | length 81 | length 81 | length 81 | length 81 |
| 8×3 | 6561 | true | length 6561 | length 6561 | length 6561 | length 6561 | length 6561 |
| 4×10 | 10000 | true | length 10000 | length 10000 | length 10000 | length 10000 | length 10000 |
| 48×2 | 281474976710656 | true | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | error: over the safety limit of 1000000 |
| 56×2 | 72057594037927936 | true | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | error: over the safety limit of 1000000 |
| 62×2 | 4611686018427387904 | true | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | panic: makeslice: len out of range | error: over the safety limit of 1000000 |
| 63×2 | 9223372036854775808 | false | panic: makeslice: len out of range | panic: makeslice: len out of range | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 64×2 | 18446744073709551616 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 65×2 | 36893488147419103232 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 21×8 | 9223372036854775808 | false | panic: makeslice: len out of range | panic: makeslice: len out of range | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 22×8 | 73786976294838206464 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 16×16 | 18446744073709551616 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 32×4 | 18446744073709551616 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 8×256 | 18446744073709551616 | false | **length 0 (wrong)** | **length 0 (wrong)** | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 7×256 + 1×128 | 9223372036854775808 | false | panic: makeslice: len out of range | panic: makeslice: len out of range | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 41×3 | 36472996377170786403 | false | panic: makeslice: len out of range | panic: makeslice: len out of range | error: result would have too many elements | error: result would have too many elements | error: over the safety limit of 1000000 |
| 64×2 + 1×0 | 0 | true | length 0 | length 0 | length 0 | length 0 | length 0 |
| 8×256 + 1×0 | 0 | true | length 0 | length 0 | length 0 | length 0 | length 0 |

Rows 48×2 to 62×2 fit in an `int` but not in any computer's memory, so every build — fixed or not — fails the allocation with a recovered panic. The fix only covers lengths an `int` cannot hold; see *What this does not show*.

`pyvider-cty` multiplies in Python integers, which cannot overflow, and refuses any product over 1,000,000 elements before building it. That limit is deliberate — its parity tracker records it as an accepted divergence from go-cty — so it answers the rows no machine can build (48×2 to 62×2) with an error instead of an allocation panic.

### 20,000 random calls, compared byte for byte

| Build | Calls | Answered | Refused | Identical to `upstream-main` |
|---|---:|---:|---:|---:|
| `upstream-v1.19.0` | 20000 | 16273 | 3727 | 20000 |
| `upstream-main` | 20000 | 16273 | 3727 | 20000 |
| `fork-equals` | 20000 | 16273 | 3727 | 20000 |
| `fork-setproduct` | 20000 | 16273 | 3727 | 20000 |
| `fork-bounds` | 20000 | 16273 | 3727 | 20000 |
| `fork-main` | 20000 | 16273 | 3727 | 20000 |
| `pyvider-cty` | — | — | — | not run |

The random calls are Go-only: they replay Go's PCG stream and compare Go renderings byte for byte, which a Python build cannot produce. pyvider-cty's own test suite compares `setproduct` with go-cty in its differential sweeps.

### Memory allocated for 6×10 (1,000,000 tuples)

| Build | Total allocated (5 runs) | Median | Result length | Result identical to `upstream-main` |
|---|---|---:|---:|---|
| `upstream-v1.19.0` | 504, 504, 504, 504, 504 MB | **504 MB** | 1000000 | true |
| `upstream-main` | 504, 504, 504, 504, 504 MB | **504 MB** | 1000000 | true |
| `fork-equals` | 504, 504, 504, 504, 504 MB | **504 MB** | 1000000 | true |
| `fork-setproduct` | 288, 288, 288, 288, 288 MB | **288 MB** | 1000000 | true |
| `fork-bounds` | 504, 504, 504, 504, 504 MB | **504 MB** | 1000000 | true |
| `fork-main` | 288, 288, 288, 288, 288 MB | **288 MB** | 1000000 | true |
| `pyvider-cty` | not measured (Python runtime) | — | 1000000 | true, by element digest |

*Element digest*: one line per tuple, its elements joined by commas, in result order — a form both harnesses compute identically, unlike Go's `%#v` rendering.

### The fix's own regression test, before and after

`TestSetproductTooManyElements`, run 20 times in one `go test -count=20`: on upstream `main` with only the fix branch's test file copied in, and on the fix branch.

| Source | Passed | Failed | Failing subtests |
|---|---:|---:|---|
| upstream `main` + test only | 0 | 20 | 63_two-element_lists<br>64_two-element_lists |
| fix branch | 20 | 0 |  |

### In real OpenTofu (`tofu console`)

| Expression | Correct | stock | built-upstream-main | built-fork-main |
|---|---|---|---|---|
| `length(setproduct(3 × ["a","b"]))` | 8 | 8 | 8 | 8 |
| `length(setproduct(8 × [256 strings]))` | 18446744073709551616 | **0 (wrong)** | **0 (wrong)** | error: Call to function "setproduct" failed: result would have too many elements. |
| `length(setproduct(7 × [256 strings], 1 × [128 strings]))` | 9223372036854775808 | error: Call to function "setproduct" failed: panic in function implementation: runtime error: makeslice: len out of range | error: Call to function "setproduct" failed: panic in function implementation: runtime error: makeslice: len out of range | error: Call to function "setproduct" failed: result would have too many elements. |

## 3. Number refinement accepts `x < x` exclusive-on-both-sides, but not one side

**Claim.** `assertConsistentBounds` picks its comparison by whether the two inclusivity flags are *equal* instead of whether both are inclusive, so `3 < x < 3` is accepted while the equally empty `3 < x ≤ 3` panics. The fix accepts equal bounds only when both are inclusive.

**Method.** Every lower and upper bound from {−1, 0, 0.5, 1, 10^38+1}, every combination of inclusive and exclusive, applied lower-first and upper-first — 200 refinements per build. The judge decides independently, with exact rational comparison, whether each range contains any number.

| Build | Refinements | Accepted | Accepted but empty | Refused but non-empty |
|---|---:|---:|---:|---:|
| `upstream-v1.19.0` | 200 | 100 | 10 | 0 |
| `upstream-main` | 200 | 100 | 10 | 0 |
| `fork-equals` | 200 | 100 | 10 | 0 |
| `fork-setproduct` | 200 | 100 | 10 | 0 |
| `fork-bounds` | 200 | 90 | 0 | 0 |
| `fork-main` | 200 | 90 | 0 | 0 |
| `pyvider-cty` | 200 | 90 | 0 | 0 |

### The ranges upstream gets wrong

| Range | Applied | `upstream-main` | `fork-bounds` | `pyvider-cty` |
|---|---|---|---|---|
| `-1 < x < -1` | lower-first | **accepted** | panics: `number lower bound cty.NumberIntVal(-1) is greater than upper bound cty.NumberIntVal(-1)` | raises `CtyRefinementError: number lower bound -1 excludes upper bound -1` |
| `-1 < x < -1` | upper-first | **accepted** | panics: `number lower bound cty.NumberIntVal(-1) is greater than upper bound cty.NumberIntVal(-1)` | raises `CtyRefinementError: number lower bound -1 excludes upper bound -1` |
| `0 < x < 0` | lower-first | **accepted** | panics: `number lower bound cty.NumberIntVal(0) is greater than upper bound cty.NumberIntVal(0)` | raises `CtyRefinementError: number lower bound 0 excludes upper bound 0` |
| `0 < x < 0` | upper-first | **accepted** | panics: `number lower bound cty.NumberIntVal(0) is greater than upper bound cty.NumberIntVal(0)` | raises `CtyRefinementError: number lower bound 0 excludes upper bound 0` |
| `0.5 < x < 0.5` | lower-first | **accepted** | panics: `number lower bound cty.NumberFloatVal(0.5) is greater than upper bound cty.NumberFloatVal(0.5)` | raises `CtyRefinementError: number lower bound 0.5 excludes upper bound 0.5` |
| `0.5 < x < 0.5` | upper-first | **accepted** | panics: `number lower bound cty.NumberFloatVal(0.5) is greater than upper bound cty.NumberFloatVal(0.5)` | raises `CtyRefinementError: number lower bound 0.5 excludes upper bound 0.5` |
| `1 < x < 1` | lower-first | **accepted** | panics: `number lower bound cty.NumberIntVal(1) is greater than upper bound cty.NumberIntVal(1)` | raises `CtyRefinementError: number lower bound 1 excludes upper bound 1` |
| `1 < x < 1` | upper-first | **accepted** | panics: `number lower bound cty.NumberIntVal(1) is greater than upper bound cty.NumberIntVal(1)` | raises `CtyRefinementError: number lower bound 1 excludes upper bound 1` |
| `100000000000000000000000000000000000001 < x < 100000000000000000000000000000000000001` | lower-first | **accepted** | panics: `number lower bound cty.NumberIntVal(1.00000000000000000000000000000000000001e+38) is greater than upper bound cty.NumberIntVal(1.00000000000000000000000000000000000001e+38)` | raises `CtyRefinementError: number lower bound 100000000000000000000000000000000000001 excludes upper bound 100000000000000000000000000000000000001` |
| `100000000000000000000000000000000000001 < x < 100000000000000000000000000000000000001` | upper-first | **accepted** | panics: `number lower bound cty.NumberIntVal(1.00000000000000000000000000000000000001e+38) is greater than upper bound cty.NumberIntVal(1.00000000000000000000000000000000000001e+38)` | raises `CtyRefinementError: number lower bound 100000000000000000000000000000000000001 excludes upper bound 100000000000000000000000000000000000001` |

### What the accepted empty range means

The value upstream accepts for `3 < x < 3`, probed:

| Build | Refinement | Probe | Result |
|---|---|---|---|
| `upstream-main` | accepted | `Range().Includes(2.5)` | `cty.False` |
| `upstream-main` | accepted | `Equals(2.5)` | `cty.False` |
| `upstream-main` | accepted | `Range().Includes(3)` | `cty.False` |
| `upstream-main` | accepted | `Equals(3)` | `cty.False` |
| `upstream-main` | accepted | `Range().Includes(3.5)` | `cty.False` |
| `upstream-main` | accepted | `Equals(3.5)` | `cty.False` |
| `upstream-main` | accepted | msgpack encoding | `c7090c82039203c2049203c2` |
| `fork-bounds` | panics: `number lower bound cty.NumberIntVal(3) is greater than upper bound cty.NumberIntVal(3)` | — | — |
| `pyvider-cty` | raises `CtyRefinementError: number lower bound 3 excludes upper bound 3` | — | — |

### Decoding refinements from the wire (`msgpack.Unmarshal`)

This is **not** part of the fix; it is recorded because the fix changes one row. go-cty's msgpack decoder applies wire bounds through the same refinement builder and does not recover its panic, so an inconsistent payload from a peer panics the decoding process. Upstream already does this for `3 < x ≤ 3` and `4 ≤ x ≤ 3`; with the fix, `3 < x < 3` behaves the same way. The payloads are hand-encoded; the `3 < x < 3` payload is byte-identical to upstream's own `msgpack.Marshal` of the value above.

| Payload | Bytes | Contains a number | `upstream-main` | `fork-bounds` | `pyvider-cty` |
|---|---|---|---|---|---|
| 3 <= x <= 4 (non-empty) | `c7090c82039203c3049204c3` | true | value | value | value |
| 3 <= x <= 3 (exactly 3) | `c7090c82039203c3049203c3` | true | value | value | value |
| 3 < x < 3 (empty, both exclusive) | `c7090c82039203c2049203c2` | false | value | panic | refused: `DeserializationError: number lower bound 3 excludes upper bound 3` |
| 3 < x <= 3 (empty, lower exclusive) | `c7090c82039203c2049203c3` | false | panic | panic | refused: `DeserializationError: number lower bound 3 excludes upper bound 3` |
| 4 <= x <= 3 (empty, lower above upper) | `c7090c82039204c3049203c3` | false | panic | panic | refused: `DeserializationError: number lower bound 4 is greater than upper bound 3` |

pyvider-cty at `v0.6.1` (`858fef516aa9`) refuses all 3 payloads whose range holds no number with `DeserializationError` — no panic and no empty range kept — and decodes the 2 that hold one.

### The fix's own regression test, before and after

`TestValueRefine`, run 20 times in one `go test -count=20`: on upstream `main` with only the fix branch's test file copied in, and on the fix branch.

| Source | Passed | Failed | Failing subtests |
|---|---:|---:|---|
| upstream `main` + test only | 0 | 20 | unknown_number_cannot_have_equal_bounds_when_both_are_exclusive |
| fix branch | 20 | 0 |  |

## Upstream's complete test suite on every fork commit

| Fork commit (branch or `main`) | Packages passed | Packages failed | `go test` exit | `go vet` exit |
|---|---:|---:|---:|---:|
| `fix/equals-unknown-and-difference` | 9 | 0 | 0 | 0 |
| `fix/setproduct-length-overflow` | 9 | 0 | 0 | 0 |
| `fix/refinement-exclusive-equal-bounds` | 9 | 0 | 0 | 0 |
| `main` | 9 | 0 | 0 | 0 |

## What this does not show

- **Equals**: operands are flat, with string members drawn from {p, q, unknown}. Nested containers, other element types and refined unknowns go through the same loops but are not enumerated here. The OpenTofu runs cover one real configuration, not Terraform itself.
- **SetProduct**: the fix adds no size limit below the range of `int`. Products that fit in `int` but not in memory (e.g. 2^40 elements) still fail the way they did before — a recovered allocation panic or an out-of-memory process — and are deliberately not run.
- **Refinement bounds**: the decoder panic on inconsistent wire payloads is pre-existing upstream behaviour, not introduced or fixed here; the fix extends it to one more payload.
- **pyvider-cty** is tested at `v0.6.1` (`858fef516aa9`) as installed from GitHub, as a library only. It has no OpenTofu column, since OpenTofu evaluates expressions with go-cty, not pyvider-cty. It is not part of the random-call replay or the allocation measurements, which are specific to Go. Its own test suite is not run here.
- The upstream issues are not filed and these branches are not proposed upstream.

