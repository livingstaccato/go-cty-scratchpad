// Command harness exercises one build of go-cty and prints what it observes as
// JSON lines. It contains no expectations: the same source is built against
// every variant (upstream releases, upstream main, and the fork's fixes), and
// the separate judge program decides what the observations prove.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	ctymsgpack "github.com/zclconf/go-cty/cty/msgpack"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: harness equals|setproduct-lengths|setproduct-diff|setproduct-memory|bounds|bounds-consequences|bounds-decode")
		os.Exit(2)
	}
	out := json.NewEncoder(os.Stdout)
	emit := func(v any) {
		if err := out.Encode(v); err != nil {
			panic(err)
		}
	}
	switch os.Args[1] {
	case "equals":
		runEquals(emit)
	case "setproduct-lengths":
		runSetProductLengths(emit)
	case "setproduct-diff":
		runSetProductDiff(emit)
	case "setproduct-memory":
		runSetProductMemory(emit)
	case "bounds":
		runBounds(emit)
	case "bounds-consequences":
		runBoundsConsequences(emit)
	case "bounds-decode":
		runBoundsDecode(emit)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", os.Args[1])
		os.Exit(2)
	}
}

var valueMarksCall = regexp.MustCompile(`cty\.NewValueMarks\(([^)]*)\)`)

// render is %#v with the marks in a fixed order. cty.ValueMarks is a map, so
// GoString lists a value's marks in Go's randomized map order: the same value
// renders as NewValueMarks("c", "e") on one run and ("e", "c") on the next.
// Without this, identical results compare as different.
func render(v any) string {
	return valueMarksCall.ReplaceAllStringFunc(fmt.Sprintf("%#v", v), func(call string) string {
		args := strings.Split(valueMarksCall.FindStringSubmatch(call)[1], ", ")
		sort.Strings(args)
		return "cty.NewValueMarks(" + strings.Join(args, ", ") + ")"
	})
}

// ---------------------------------------------------------------------------
// Value.Equals
// ---------------------------------------------------------------------------

// Each position of each operand is one of: the known string "p", the known
// string "q", or an unknown string ("U").
var equalsCells = []byte{'p', 'q', 'U'}

// A comparison is repeated this many times, because object attributes and map
// elements are visited in Go's randomized iteration order.
const equalsIterations = 200

// Values substituted for an unknown position when computing the ground truth.
// "r" is a value that appears in neither operand.
var equalsSubstitutes = []byte{'p', 'q', 'r'}

type equalsObservation struct {
	Category string `json:"category"` // "true", "false" or "unknown"
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

func cellStrings(n int) []string {
	if n == 0 {
		return []string{""}
	}
	var ret []string
	for _, prefix := range cellStrings(n - 1) {
		for _, c := range equalsCells {
			ret = append(ret, prefix+string(c))
		}
	}
	return ret
}

func equalsOperand(kind, cells string, marked bool) cty.Value {
	vals := make([]cty.Value, len(cells))
	for i := range len(cells) {
		if cells[i] == 'U' {
			vals[i] = cty.UnknownVal(cty.String)
		} else {
			vals[i] = cty.StringVal(string(cells[i]))
		}
	}
	if marked {
		vals[0] = vals[0].Mark("sensitive")
	}
	switch kind {
	case "object", "map":
		attrs := make(map[string]cty.Value, len(vals))
		for i, v := range vals {
			attrs[fmt.Sprintf("k%d", i)] = v
		}
		if kind == "object" {
			return cty.ObjectVal(attrs)
		}
		return cty.MapVal(attrs)
	case "list":
		return cty.ListVal(vals)
	case "tuple":
		return cty.TupleVal(vals)
	}
	panic("unsupported kind " + kind)
}

// equalsTruth substitutes every combination of concrete values for the unknown
// positions and reports which equality results are possible.
func equalsTruth(lhs, rhs string) []string {
	joined := []byte(lhs + rhs)
	var unknowns []int
	for i, c := range joined {
		if c == 'U' {
			unknowns = append(unknowns, i)
		}
	}
	seen := map[string]bool{}
	combos := int(math.Pow(float64(len(equalsSubstitutes)), float64(len(unknowns))))
	for combo := range combos {
		concrete := append([]byte(nil), joined...)
		rest := combo
		for _, pos := range unknowns {
			concrete[pos] = equalsSubstitutes[rest%len(equalsSubstitutes)]
			rest /= len(equalsSubstitutes)
		}
		seen[fmt.Sprint(string(concrete[:len(lhs)]) == string(concrete[len(lhs):]))] = true
	}
	var ret []string
	for k := range seen {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

func category(v cty.Value) string {
	v, _ = v.Unmark()
	switch {
	case !v.IsKnown():
		return "unknown"
	case v.True():
		return "true"
	default:
		return "false"
	}
}

func runEquals(emit func(any)) {
	for _, kind := range []string{"object", "map", "list", "tuple"} {
		for n := 1; n <= 4; n++ {
			for _, lhs := range cellStrings(n) {
				for _, rhs := range cellStrings(n) {
					for _, marked := range []bool{false, true} {
						if marked && n > 3 {
							continue
						}
						l, r := equalsOperand(kind, lhs, marked), equalsOperand(kind, rhs, false)
						seen := map[string]string{}
						for range equalsIterations {
							got := l.Equals(r)
							seen[render(got)] = category(got)
						}
						c := equalsCase{Kind: kind, LHS: lhs, RHS: rhs, Marked: marked, Truth: equalsTruth(lhs, rhs)}
						for gs, cat := range seen {
							c.Observed = append(c.Observed, equalsObservation{Category: cat, GoString: gs})
						}
						sort.Slice(c.Observed, func(i, j int) bool { return c.Observed[i].GoString < c.Observed[j].GoString })
						emit(c)
					}
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// SetProductFunc
// ---------------------------------------------------------------------------

type lengthGroup struct {
	Count  int `json:"count"`
	Length int `json:"length"`
}

type lengthCase struct {
	Groups       []lengthGroup `json:"groups"`
	Expected     string        `json:"expected"`
	FitsInt      bool          `json:"fits_int"`
	WrappedInt   int64         `json:"wrapped_int"`
	Outcome      string        `json:"outcome"` // "ok", "error" or "skipped"
	Length       int           `json:"length"`
	Error        string        `json:"error,omitempty"`
	ResultSHA256 string        `json:"result_sha256,omitempty"`
	ValuesSHA256 string        `json:"values_sha256,omitempty"`
}

// valuesDigest hashes a product of string lists in a form any language can
// reproduce: one line per tuple, its elements joined by commas, in result
// order. The pyvider-cty harness computes the same digest.
func valuesDigest(v cty.Value) string {
	h := sha256.New()
	for it := v.ElementIterator(); it.Next(); {
		_, tuple := it.Element()
		var parts []string
		for _, e := range tuple.AsValueSlice() {
			parts = append(parts, e.AsString())
		}
		fmt.Fprintf(h, "%s\n", strings.Join(parts, ","))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func collection(length int) cty.Value {
	if length == 0 {
		return cty.ListValEmpty(cty.String)
	}
	elems := make([]cty.Value, length)
	for i := range elems {
		elems[i] = cty.StringVal(fmt.Sprint(i))
	}
	return cty.ListVal(elems)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func runSetProductLengths(emit func(any)) {
	var specs [][]lengthGroup
	for n := 2; n <= 16; n++ {
		specs = append(specs, []lengthGroup{{n, 2}})
	}
	for n := 2; n <= 8; n++ {
		specs = append(specs, []lengthGroup{{n, 3}})
	}
	specs = append(specs,
		[]lengthGroup{{4, 10}},
		[]lengthGroup{{48, 2}},
		[]lengthGroup{{56, 2}},
		[]lengthGroup{{62, 2}},
		[]lengthGroup{{63, 2}},
		[]lengthGroup{{64, 2}},
		[]lengthGroup{{65, 2}},
		[]lengthGroup{{21, 8}},
		[]lengthGroup{{22, 8}},
		[]lengthGroup{{16, 16}},
		[]lengthGroup{{32, 4}},
		[]lengthGroup{{8, 256}},
		[]lengthGroup{{7, 256}, {1, 128}},
		[]lengthGroup{{41, 3}},
		[]lengthGroup{{64, 2}, {1, 0}},
		[]lengthGroup{{8, 256}, {1, 0}},
	)

	for _, groups := range specs {
		expected := big.NewInt(1)
		var wrapped uint64 = 1
		var args []cty.Value
		for _, g := range groups {
			for range g.Count {
				expected.Mul(expected, big.NewInt(int64(g.Length)))
				wrapped *= uint64(g.Length)
				args = append(args, collection(g.Length))
			}
		}
		c := lengthCase{
			Groups:     groups,
			Expected:   expected.String(),
			FitsInt:    expected.Cmp(big.NewInt(math.MaxInt)) <= 0,
			WrappedInt: int64(wrapped),
		}
		// An unfixed build allocates whatever the wrapped length says. A
		// wrapped length in this band is large enough to exhaust memory but too
		// small for the runtime to refuse outright, so it is not attempted.
		if c.WrappedInt > 1_000_000 && c.WrappedInt < 1<<40 {
			c.Outcome = "skipped"
			emit(c)
			continue
		}
		got, err := stdlib.SetProductFunc.Call(args)
		if err != nil {
			c.Outcome, c.Error = "error", firstLine(err.Error())
		} else {
			c.Outcome, c.Length = "ok", got.LengthInt()
			if c.Length <= 100_000 {
				sum := sha256.Sum256([]byte(render(got)))
				c.ResultSHA256 = hex.EncodeToString(sum[:])
				c.ValuesSHA256 = valuesDigest(got)
			}
		}
		emit(c)
	}
}

type diffCase struct {
	Index   int    `json:"index"`
	Call    string `json:"call"`
	Result  string `json:"result"`
	SHA256  string `json:"sha256"`
	Outcome string `json:"outcome"`
}

// diffArg builds one random argument. It returns ok=false when cty itself
// refuses to construct the value, in which case the caller draws again.
func diffArg(r *rand.Rand) (v cty.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	etys := []cty.Type{cty.String, cty.Number, cty.Bool}
	elem := func(ty cty.Type) cty.Value {
		var v cty.Value
		switch {
		case r.IntN(8) == 0:
			v = cty.UnknownVal(ty)
		case ty == cty.String:
			v = cty.StringVal([]string{"a", "b", "c"}[r.IntN(3)])
		case ty == cty.Number:
			v = cty.NumberIntVal(int64(r.IntN(3)))
		default:
			v = cty.BoolVal(r.IntN(2) == 0)
		}
		if r.IntN(10) == 0 {
			v = v.Mark("e")
		}
		return v
	}
	var ret cty.Value
	switch r.IntN(10) {
	case 0, 1, 2:
		ety := etys[r.IntN(len(etys))]
		n := r.IntN(4)
		if n == 0 {
			ret = cty.ListValEmpty(ety)
			break
		}
		elems := make([]cty.Value, n)
		for i := range elems {
			elems[i] = elem(ety)
		}
		ret = cty.ListVal(elems)
	case 3, 4, 5:
		ety := etys[r.IntN(len(etys))]
		n := r.IntN(4)
		if n == 0 {
			ret = cty.SetValEmpty(ety)
			break
		}
		elems := make([]cty.Value, n)
		for i := range elems {
			elems[i] = elem(ety)
		}
		ret = cty.SetVal(elems)
	case 6, 7:
		n := r.IntN(4)
		elems := make([]cty.Value, n)
		for i := range elems {
			elems[i] = elem(etys[r.IntN(len(etys))])
		}
		ret = cty.TupleVal(elems)
	case 8:
		ret = cty.UnknownVal(cty.List(cty.String))
		if r.IntN(2) == 0 {
			ret = ret.Refine().CollectionLengthLowerBound(1).CollectionLengthUpperBound(3).NewValue()
		}
	default:
		if r.IntN(2) == 0 {
			ret = cty.DynamicVal
		} else {
			ret = cty.NullVal(cty.List(cty.String))
		}
	}
	if r.IntN(7) == 0 {
		ret = ret.Mark("c")
	}
	return ret, true
}

func runSetProductDiff(emit func(any)) {
	r := rand.New(rand.NewPCG(20260912, 221))
	for i := range 20000 {
		argc := 2 + r.IntN(3)
		args := make([]cty.Value, 0, argc)
		for len(args) < argc {
			if v, ok := diffArg(r); ok {
				args = append(args, v)
			}
		}
		call := render(args)
		var result, outcome string
		got, err := stdlib.SetProductFunc.Call(args)
		if err != nil {
			outcome, result = "error", firstLine(err.Error())
		} else {
			outcome, result = "ok", render(got)
		}
		sum := sha256.Sum256([]byte(result))
		emit(diffCase{Index: i, Call: call, Result: result, SHA256: hex.EncodeToString(sum[:]), Outcome: outcome})
	}
}

type memoryCase struct {
	Samples      []uint64 `json:"total_alloc_bytes"`
	Length       int      `json:"length"`
	ResultSHA256 string   `json:"result_sha256"`
	ValuesSHA256 string   `json:"values_sha256"`
}

func runSetProductMemory(emit func(any)) {
	args := make([]cty.Value, 6)
	for i := range args {
		args[i] = collection(10)
	}
	var c memoryCase
	var last cty.Value
	for range 5 {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		got, err := stdlib.SetProductFunc.Call(args)
		runtime.ReadMemStats(&after)
		if err != nil {
			panic(err)
		}
		c.Samples = append(c.Samples, after.TotalAlloc-before.TotalAlloc)
		last = got
	}
	c.Length = last.LengthInt()
	h := sha256.New()
	fmt.Fprintf(h, "%#v", last)
	c.ResultSHA256 = hex.EncodeToString(h.Sum(nil))
	c.ValuesSHA256 = valuesDigest(last)
	emit(c)
}

// ---------------------------------------------------------------------------
// Number refinement bounds
// ---------------------------------------------------------------------------

var boundValues = []string{"-1", "0", "0.5", "1", "100000000000000000000000000000000000001"}

type boundsCase struct {
	Lower    string `json:"lower"`
	Upper    string `json:"upper"`
	LowerInc bool   `json:"lower_inclusive"`
	UpperInc bool   `json:"upper_inclusive"`
	Order    string `json:"order"`
	Accepted bool   `json:"accepted"`
	Panic    string `json:"panic,omitempty"`
	GoString string `json:"gostring,omitempty"`
}

func refineBounds(lower, upper string, lowerInc, upperInc, lowerFirst bool) (ret cty.Value, panicMsg string) {
	defer func() {
		if p := recover(); p != nil {
			panicMsg = fmt.Sprint(p)
		}
	}()
	lo, hi := cty.MustParseNumberVal(lower), cty.MustParseNumberVal(upper)
	b := cty.UnknownVal(cty.Number).Refine()
	if lowerFirst {
		b = b.NumberRangeLowerBound(lo, lowerInc).NumberRangeUpperBound(hi, upperInc)
	} else {
		b = b.NumberRangeUpperBound(hi, upperInc).NumberRangeLowerBound(lo, lowerInc)
	}
	return b.NewValue(), ""
}

func runBounds(emit func(any)) {
	for _, lower := range boundValues {
		for _, upper := range boundValues {
			for _, lowerInc := range []bool{false, true} {
				for _, upperInc := range []bool{false, true} {
					for _, lowerFirst := range []bool{true, false} {
						c := boundsCase{Lower: lower, Upper: upper, LowerInc: lowerInc, UpperInc: upperInc, Order: "upper-first"}
						if lowerFirst {
							c.Order = "lower-first"
						}
						v, p := refineBounds(lower, upper, lowerInc, upperInc, lowerFirst)
						if p != "" {
							c.Panic = p
						} else {
							c.Accepted, c.GoString = true, fmt.Sprintf("%#v", v)
						}
						emit(c)
					}
				}
			}
		}
	}
}

type consequence struct {
	Probe  string `json:"probe"`
	Result string `json:"result"`
}

type consequencesCase struct {
	Accepted     bool          `json:"accepted"`
	Panic        string        `json:"panic,omitempty"`
	GoString     string        `json:"gostring,omitempty"`
	Consequences []consequence `json:"consequences,omitempty"`
	MsgpackHex   string        `json:"msgpack_hex,omitempty"`
}

func runBoundsConsequences(emit func(any)) {
	v, p := refineBounds("3", "3", false, false, true)
	c := consequencesCase{Accepted: p == "", Panic: p}
	if p == "" {
		c.GoString = fmt.Sprintf("%#v", v)
		rng := v.Range()
		for _, n := range []string{"2.5", "3", "3.5"} {
			num := cty.MustParseNumberVal(n)
			c.Consequences = append(c.Consequences,
				consequence{Probe: "Range().Includes(" + n + ")", Result: fmt.Sprintf("%#v", rng.Includes(num))},
				consequence{Probe: "Equals(" + n + ")", Result: fmt.Sprintf("%#v", v.Equals(num))},
			)
		}
		b, err := ctymsgpack.Marshal(v, cty.Number)
		if err != nil {
			c.MsgpackHex = "error: " + err.Error()
		} else {
			c.MsgpackHex = hex.EncodeToString(b)
		}
	}
	emit(c)
}

type decodeCase struct {
	Description string `json:"description"`
	PayloadHex  string `json:"payload_hex"`
	Outcome     string `json:"outcome"` // "value", "error" or "panic"
	Detail      string `json:"detail"`
}

func runBoundsDecode(emit func(any)) {
	// A refined unknown number as msgpack extension 12, whose body is a map of
	// refinement keys: 3 is the lower bound and 4 the upper, each a
	// [number, inclusive] array.
	payload := func(lower int, lowerInc bool, upper int, upperInc bool) string {
		flag := func(b bool) string {
			if b {
				return "c3"
			}
			return "c2"
		}
		body := fmt.Sprintf("82039%d%02x%s04%s%02x%s", 2, lower, flag(lowerInc), "92", upper, flag(upperInc))
		return fmt.Sprintf("c7%02x0c%s", len(body)/2, body)
	}
	cases := []struct {
		desc string
		hex  string
	}{
		{"3 <= x <= 4 (non-empty)", payload(3, true, 4, true)},
		{"3 <= x <= 3 (exactly 3)", payload(3, true, 3, true)},
		{"3 < x < 3 (empty, both exclusive)", payload(3, false, 3, false)},
		{"3 < x <= 3 (empty, lower exclusive)", payload(3, false, 3, true)},
		{"4 <= x <= 3 (empty, lower above upper)", payload(4, true, 3, true)},
	}
	for _, tc := range cases {
		c := decodeCase{Description: tc.desc, PayloadHex: tc.hex}
		raw, err := hex.DecodeString(tc.hex)
		if err != nil {
			panic(err)
		}
		func() {
			defer func() {
				if p := recover(); p != nil {
					c.Outcome, c.Detail = "panic", fmt.Sprint(p)
				}
			}()
			v, err := ctymsgpack.Unmarshal(raw, cty.Number)
			if err != nil {
				c.Outcome, c.Detail = "error", err.Error()
				return
			}
			c.Outcome, c.Detail = "value", fmt.Sprintf("%#v", v)
		}()
		emit(c)
	}
}
