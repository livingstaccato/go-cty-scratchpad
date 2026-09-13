#!/usr/bin/env bash
# Regenerates every piece of evidence behind REPORT.md, from nothing.
#
# Every go-cty build is fetched through the Go module proxy with its checksum
# verified by sum.golang.org, so no result depends on a local copy of go-cty
# that could have been edited:
#
#   upstream-v1.19.0   github.com/zclconf/go-cty@v1.19.0 (release)
#   upstream-main      github.com/zclconf/go-cty@<pinned upstream main commit>
#   fork-*             github.com/livingstaccato/go-cty@<pinned commit>, via replace
#
#   pyvider-cty        github.com/provide-io/pyvider-cty@<pinned tag>, via uv pip
#
# The two sources read as git trees -- the fork, for its regression tests and
# full suites, and OpenTofu -- are cloned over https into build/ on first use.
# Nothing depends on where this directory lives or what it is called.
#
# Usage: ./run-proof.sh [prs|variants|pyvider|unit-tests|tofu-stock|tofu-built|report|all]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
RESULTS="$ROOT/results"
BUILD="$ROOT/build"
PLAN_RUNS="${PLAN_RUNS:-30}"

# Where everything comes from.
UPSTREAM_MODULE="github.com/zclconf/go-cty"
UPSTREAM_BASE="a918e1174fcf2a25b7a222e7e78b00ea40ace26c"
FORK_REPO="livingstaccato/go-cty"
FORK_MODULE="github.com/$FORK_REPO"
FORK_URL="https://github.com/$FORK_REPO.git"
FORK_CLONE="${FORK_CLONE:-$BUILD/go-cty}"
OPENTOFU_URL="https://github.com/opentofu/opentofu.git"
OPENTOFU_TAG="v1.12.6"
OPENTOFU_COMMIT="b4305e5a5dd2fb79a27897ae30784a181d3a26cb"
OPENTOFU_SRC="$BUILD/opentofu"
PYVIDER_URL="https://github.com/provide-io/pyvider-cty.git"
PYVIDER_REF="${PYVIDER_REF:-v0.6.1}"
PYTHON_VERSION="3.13"

# The oldest Go the harness builds with, and the OpenTofu the stock runs use.
GO_MIN="1.25"
TOFU_VERSION="1.12.6"

# The commit each fork variant is built at, all of them on the fork's `main`, so
# they stay reachable. Each fix is pinned to the commit that carries that fix
# alone: its PR head, or, where the head only merged `main` into the branch,
# that head's first parent. `fork-main` is where all three pull requests merged.
FORK_EQUALS_COMMIT="ef2c1cac99f51a4ad01d9f94d5bec410efd6ea81"
FORK_SETPRODUCT_COMMIT="e7c9522b1c2014ccbf6c2502a50f4c83a9b26c3c"
FORK_BOUNDS_COMMIT="781f24c6bf5e768fa45971d8ab217ad4a120df06"
FORK_MAIN_COMMIT="d8ae3d7fbbe2a28474d4b4edb834b8a660056ae6"

# name|source|ref -- a fork ref is the branch a commit was proposed from, kept
# as its label so the report can match it to that pull request.
VARIANTS=(
  "upstream-v1.19.0|upstream|v1.19.0"
  "upstream-main|upstream|$UPSTREAM_BASE"
  "fork-equals|fork|fix/equals-unknown-and-difference"
  "fork-setproduct|fork|fix/setproduct-length-overflow"
  "fork-bounds|fork|fix/refinement-exclusive-equal-bounds"
  "fork-main|fork|main"
)

HARNESS_COMMANDS=(equals setproduct-lengths setproduct-diff setproduct-memory bounds bounds-consequences bounds-decode)
PYHARNESS_COMMANDS=(equals setproduct-lengths setproduct-memory bounds bounds-consequences bounds-decode)

fork_commit() {
  case "$1" in
    fix/equals-unknown-and-difference) echo "$FORK_EQUALS_COMMIT" ;;
    fix/setproduct-length-overflow) echo "$FORK_SETPRODUCT_COMMIT" ;;
    fix/refinement-exclusive-equal-bounds) echo "$FORK_BOUNDS_COMMIT" ;;
    main) echo "$FORK_MAIN_COMMIT" ;;
    *) echo "no pinned commit for $1" >&2; return 1 ;;
  esac
}

# Fails early, naming what is missing, rather than partway through a long run.
preflight() {
  local missing=() tool go_version tofu_line
  for tool in git go gh jq uv tofu tar awk; do
    command -v "$tool" > /dev/null 2>&1 || missing+=("$tool")
  done
  [ "${#missing[@]}" = 0 ] || { echo "missing tools: ${missing[*]}" >&2; exit 1; }
  go_version="$(go env GOVERSION)"
  go_version="${go_version#go}"
  [ "$(printf '%s\n%s\n' "$GO_MIN" "$go_version" | sort -V | head -n 1)" = "$GO_MIN" ] \
    || { echo "Go $GO_MIN or newer is required, found $go_version" >&2; exit 1; }
  tofu_line="$(tofu version)"
  tofu_line="${tofu_line%%$'\n'*}"
  [ "$tofu_line" = "OpenTofu v$TOFU_VERSION" ] \
    || { echo "OpenTofu $TOFU_VERSION is required on PATH, found: $tofu_line" >&2; exit 1; }
  gh api rate_limit --silent > /dev/null 2>&1 || { echo "gh cannot reach GitHub: run 'gh auth login' or set GH_TOKEN" >&2; exit 1; }
}

# The fork's git history, for the steps that build from source rather than
# through the module proxy.
ensure_fork_clone() {
  [ -d "$FORK_CLONE/.git" ] || git clone -q "$FORK_URL" "$FORK_CLONE"
  git -C "$FORK_CLONE" fetch -q origin
}

ensure_opentofu() {
  [ -d "$OPENTOFU_SRC/.git" ] || git clone -q --depth 1 --branch "$OPENTOFU_TAG" "$OPENTOFU_URL" "$OPENTOFU_SRC"
  [ "$(git -C "$OPENTOFU_SRC" rev-parse HEAD)" = "$OPENTOFU_COMMIT" ] \
    || { echo "$OPENTOFU_SRC is not OpenTofu $OPENTOFU_TAG ($OPENTOFU_COMMIT)" >&2; exit 1; }
}

build_variant() {
  local name="$1" source="$2" ref="$3" dir="$BUILD/$1" sha
  rm -rf "$dir" && mkdir -p "$dir" "$RESULTS/$name"
  cp "$ROOT/proof/harness/main.go" "$dir/"
  printf 'module proof\n\ngo %s\n' "$GO_MIN" > "$dir/go.mod"
  (
    cd "$dir"
    if [ "$source" = upstream ]; then
      go get "$UPSTREAM_MODULE@$ref" >/dev/null 2>&1
      sha="$ref"
    else
      sha="$(fork_commit "$ref")"
      go get "$UPSTREAM_MODULE@v1.19.0" >/dev/null 2>&1
      go mod edit -replace "$UPSTREAM_MODULE=$FORK_MODULE@$sha"
    fi
    go mod tidy >/dev/null 2>&1
    go mod verify >/dev/null
    go build -o harness .
    go list -m -json "$UPSTREAM_MODULE" > "$RESULTS/$name/module.json"
    grep -E '(zclconf|livingstaccato)/go-cty ' go.sum > "$RESULTS/$name/go.sum.lines"
    printf '{"name":"%s","source":"%s","ref":"%s","commit":"%s","go":"%s"}\n' \
      "$name" "$source" "$ref" "$sha" "$(go env GOVERSION)" > "$RESULTS/$name/variant.json"
  )
}

run_variants() {
  for spec in "${VARIANTS[@]}"; do
    IFS='|' read -r name source ref <<< "$spec"
    echo "== building $name ($source $ref)"
    build_variant "$name" "$source" "$ref"
    for cmd in "${HARNESS_COMMANDS[@]}"; do
      echo "   running $cmd"
      "$BUILD/$name/harness" "$cmd" > "$RESULTS/$name/$cmd.jsonl"
    done
  done
}

# The same cases against pyvider-cty, installed from GitHub at a pinned tag,
# never from a local checkout. The installed commit is read back from the
# package's own direct_url.json, so the report can show it is the commit GitHub
# has for that tag.
run_pyvider() {
  local dir="$BUILD/pyvider-cty" out="$RESULTS/pyvider-cty" refs sha
  refs="$(git ls-remote "$PYVIDER_URL" "refs/tags/$PYVIDER_REF" "refs/heads/$PYVIDER_REF")"
  sha="$(printf '%s\n' "$refs" | awk '/\^\{\}$/ {peeled = $1} NR == 1 {first = $1} END {print (peeled != "" ? peeled : first)}')"
  [ -n "$sha" ] || { echo "$PYVIDER_REF is not a tag or branch on $PYVIDER_URL" >&2; exit 1; }
  echo "== installing pyvider-cty $PYVIDER_REF ($sha)"
  rm -rf "$dir" "$out" && mkdir -p "$dir" "$out"
  UV_NO_SOURCES=1 uv venv -q --python "$PYTHON_VERSION" "$dir/.venv"
  UV_NO_SOURCES=1 uv pip install -q --python "$dir/.venv/bin/python" "pyvider-cty @ git+$PYVIDER_URL@$sha"
  "$dir/.venv/bin/python" - "$PYVIDER_REF" "$sha" > "$out/variant.json" <<'PY'
import importlib.metadata, json, platform, sys
dist = importlib.metadata.distribution("pyvider-cty")
installed = json.loads(dist.read_text("direct_url.json") or "{}").get("vcs_info", {}).get("commit_id", "")
print(json.dumps({"ref": sys.argv[1], "commit": sys.argv[2], "installed_commit": installed,
                  "version": dist.version, "python": platform.python_version()}))
PY
  for cmd in "${PYHARNESS_COMMANDS[@]}"; do
    echo "   running $cmd"
    "$dir/.venv/bin/python" "$ROOT/proof/pyharness/harness.py" "$cmd" > "$out/$cmd.jsonl"
  done
}

# Each fix's own regression test, run twice: once on the unfixed upstream
# source with only that test file swapped in (it must fail), and once on the
# fix commit (it must pass). Then the full upstream suite on every fork commit.
run_unit_tests() {
  local out="$RESULTS/unit-tests" work="$ROOT/work/unit-tests"
  rm -rf "$out" "$work" && mkdir -p "$out" "$work"
  local specs=(
    "equals|fix/equals-unknown-and-difference|cty/value_ops_test.go|./cty|^TestValueEqualsUnknownAndDifference\$"
    "setproduct|fix/setproduct-length-overflow|cty/function/stdlib/collection_test.go|./cty/function/stdlib|^TestSetproductTooManyElements\$"
    "bounds|fix/refinement-exclusive-equal-bounds|cty/unknown_refinement_test.go|./cty|^TestValueRefine\$"
  )
  ensure_fork_clone
  for spec in "${specs[@]}"; do
    IFS='|' read -r name branch testfile pkg pattern <<< "$spec"
    local sha; sha="$(fork_commit "$branch")"
    echo "== regression test for $name"
    mkdir -p "$work/$name-base" "$work/$name-fix"
    git -C "$FORK_CLONE" archive "$UPSTREAM_BASE" | tar -x -C "$work/$name-base"
    git -C "$FORK_CLONE" show "$sha:$testfile" > "$work/$name-base/$testfile"
    git -C "$FORK_CLONE" archive "$sha" | tar -x -C "$work/$name-fix"
    for side in base fix; do
      (cd "$work/$name-$side" && go test "$pkg" -run "$pattern" -count=20 -json > "$out/$name-$side.json" 2>&1) || true
    done
  done
  for branch in fix/equals-unknown-and-difference fix/setproduct-length-overflow fix/refinement-exclusive-equal-bounds main; do
    local sha slug; sha="$(fork_commit "$branch")"; slug="${branch//\//-}"
    echo "== full suite on $branch ($sha)"
    mkdir -p "$work/full-$slug"
    git -C "$FORK_CLONE" archive "$sha" | tar -x -C "$work/full-$slug"
    (cd "$work/full-$slug" && go test ./... -count=1 -json > "$out/full-$slug.json" 2>&1; echo $? > "$out/full-$slug.exit") || true
    (cd "$work/full-$slug" && go vet ./... > "$out/vet-$slug.txt" 2>&1; echo $? > "$out/vet-$slug.exit") || true
  done
}

# tofu_runs <label> <tofu binary>: the same three configurations, run PLAN_RUNS
# times each, plus the setproduct expressions through `tofu console`.
tofu_runs() {
  local tofu="$2" out="$RESULTS/tofu/$1"
  rm -rf "$out" && mkdir -p "$out"
  "$tofu" version > "$out/version.txt" 2>&1
  for config in equals count; do
    (cd "$ROOT/tofu/$config" && rm -rf .terraform .terraform.lock.hcl && "$tofu" init -no-color > /dev/null 2>&1)
  done
  for _ in $(seq 1 "$PLAN_RUNS"); do
    (cd "$ROOT/tofu/equals" && "$tofu" plan -no-color 2>&1 | grep -E '^\s+\+ (objects|maps)_equal' | tr -s ' ' | sed 's/^ *//') >> "$out/equals-plans.txt"
    if (cd "$ROOT/tofu/count" && "$tofu" plan -no-color > "$out/count-plan.tmp" 2>&1); then
      echo "succeeded: $(grep -E '^Plan:' "$out/count-plan.tmp")" >> "$out/count-plans.txt"
    else
      echo "failed: $(grep -m1 -E '^Error:' "$out/count-plan.tmp")" >> "$out/count-plans.txt"
    fi
  done
  rm -f "$out/count-plan.tmp"
  local exprs=(
    'length(setproduct([for i in range(3) : ["a", "b"]]...))'
    'length(setproduct([for i in range(8) : [for j in range(256) : tostring(j)]]...))'
    'length(setproduct(concat([for i in range(7) : [for j in range(256) : tostring(j)]], [[for j in range(128) : tostring(j)]])...))'
  )
  mkdir -p "$ROOT/tofu/console"
  for e in "${exprs[@]}"; do
    printf '> %s\n' "$e" >> "$out/console.txt"
    (cd "$ROOT/tofu/console" && echo "$e" | "$tofu" console -no-color >> "$out/console.txt" 2>&1) || true
  done
}

run_tofu_stock() {
  tofu_runs stock "$(command -v tofu)"
}

# OpenTofu built from source twice, identical except for go-cty: once on
# upstream main, once on the fork's main.
run_tofu_built() {
  ensure_opentofu
  for spec in "upstream-main|upstream|$UPSTREAM_BASE" "fork-main|fork|main"; do
    IFS='|' read -r name source ref <<< "$spec"
    local dir="$BUILD/tofu-$name" sha
    rm -rf "$dir" && mkdir -p "$dir"
    git -C "$OPENTOFU_SRC" archive HEAD | tar -x -C "$dir"
    (
      cd "$dir"
      # OpenTofu pins its Go version in go.mod, and newer Go breaks its pinned
      # golang.org/x/net. Build with exactly the toolchain it names.
      GOTOOLCHAIN="go$(awk '$1 == "go" {print $2}' go.mod)"
      export GOTOOLCHAIN
      if [ "$source" = upstream ]; then
        go get "$UPSTREAM_MODULE@$ref" >/dev/null 2>&1
      else
        sha="$(fork_commit "$ref")"
        go mod edit -replace "$UPSTREAM_MODULE=$FORK_MODULE@$sha"
      fi
      go mod tidy >/dev/null 2>&1
      go build -o tofu ./cmd/tofu
    )
    go version -m "$dir/tofu" | grep -E 'go-cty' > "$RESULTS/tofu-built-$name.modules.txt"
    tofu_runs "built-$name" "$dir/tofu"
  done
}

# The fork's pull requests as GitHub reports them, so the report can check that
# each PR's head is the commit that was tested, or only merges its base into it.
# For each head: its parents, and whether a second parent is on the base branch
# (GitHub's compare reports it "behind" or "identical" to the base).
run_prs() {
  mkdir -p "$RESULTS"
  gh pr list -R "$FORK_REPO" --state all \
    --json number,title,url,headRefName,headRefOid,baseRefName,state > "$RESULTS/prs.json"
  jq -r '.[] | "\(.headRefOid) \(.baseRefName)"' "$RESULTS/prs.json" | while read -r head base; do
    on_base=false
    parents="$(gh api "repos/$FORK_REPO/commits/$head" --jq '[.parents[].sha] | join(" ")')"
    read -r -a list <<< "$parents"
    if [ "${#list[@]}" = 2 ]; then
      case "$(gh api "repos/$FORK_REPO/compare/$base...${list[1]}" --jq .status)" in
        behind|identical) on_base=true ;;
      esac
    fi
    jq -n --arg head "$head" --arg parents "$parents" --argjson on_base "$on_base" \
      '{head: $head, parents: ($parents | split(" ")), second_parent_on_base: $on_base}'
  done | jq -s . > "$RESULTS/prs-parents.json"
}

run_report() {
  (cd "$ROOT/proof/judge" && go run . "$RESULTS" "$ROOT/REPORT.md")
}

step="${1:-all}"
case "$step" in
  prs|variants|pyvider|unit-tests|tofu-stock|tofu-built|all) preflight; mkdir -p "$BUILD" ;;
esac
case "$step" in
  variants) run_variants ;;
  pyvider) run_pyvider ;;
  unit-tests) run_unit_tests ;;
  tofu-stock) run_tofu_stock ;;
  tofu-built) run_tofu_built ;;
  prs) run_prs ;;
  report) run_report ;;
  all) rm -rf "$RESULTS"; run_prs; run_variants; run_pyvider; run_unit_tests; run_tofu_stock; run_tofu_built; run_report ;;
  *) echo "usage: $0 [prs|variants|pyvider|unit-tests|tofu-stock|tofu-built|report|all]" >&2; exit 2 ;;
esac
