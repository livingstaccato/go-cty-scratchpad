"""pyvider-cty's side of the proof: the same cases proof/harness/main.go runs
against go-cty, run against pyvider-cty, printed as JSON lines in the same
shapes. Like the Go harness it contains no expectations; the judge grades it.

Usage: python harness.py equals|setproduct-lengths|setproduct-memory|bounds|bounds-consequences|bounds-decode
"""

from __future__ import annotations

from collections.abc import Callable, Iterator
from decimal import Decimal
import hashlib
import itertools
import json
import os
import sys
from typing import Any

from pyvider.cty import CtyList, CtyMap, CtyNumber, CtyObject, CtyString, CtyTuple, CtyValue
from pyvider.cty.codec import cty_from_msgpack
from pyvider.cty.functions import setproduct
from pyvider.cty.refinement import refine
from pyvider.cty.value_range import value_range

Emit = Callable[[Any], None]

STRING = CtyString()
NUMBER = CtyNumber()


def emit(obj: Any) -> None:
    sys.stdout.write(json.dumps(obj, separators=(",", ":"), ensure_ascii=False) + "\n")


def described(exc: BaseException) -> str:
    return f"{type(exc).__name__}: {str(exc).splitlines()[0] if str(exc) else ''}"


# ---------------------------------------------------------------------------
# Value.Equals
# ---------------------------------------------------------------------------

EQUALS_CELLS = "pqU"
EQUALS_SUBSTITUTES = "pqr"
# The Go harness repeats each comparison because Go randomizes map iteration.
# Python has no such mechanism, but the count is kept so both are held to the
# same test.
EQUALS_ITERATIONS = int(os.environ.get("PYHARNESS_EQUALS_ITERATIONS", "200"))


def cell_strings(n: int) -> Iterator[str]:
    return ("".join(p) for p in itertools.product(EQUALS_CELLS, repeat=n))


def equals_operand(kind: str, cells: str, marked: bool) -> CtyValue[Any]:
    vals = [CtyValue.unknown(STRING) if c == "U" else STRING.validate(c) for c in cells]
    if marked:
        vals[0] = vals[0].mark("sensitive")
    if kind in ("object", "map"):
        attrs = {f"k{i}": v for i, v in enumerate(vals)}
        if kind == "object":
            return CtyObject(attribute_types=dict.fromkeys(attrs, STRING)).validate(attrs)
        return CtyMap(element_type=STRING).validate(attrs)
    if kind == "list":
        return CtyList(element_type=STRING).validate(vals)
    if kind == "tuple":
        return CtyTuple(element_types=tuple(STRING for _ in vals)).validate(vals)
    raise ValueError(kind)


def equals_truth(lhs: str, rhs: str) -> list[str]:
    joined = lhs + rhs
    unknowns = [i for i, c in enumerate(joined) if c == "U"]
    seen = set()
    for combo in itertools.product(EQUALS_SUBSTITUTES, repeat=len(unknowns)):
        concrete = list(joined)
        for pos, sub in zip(unknowns, combo, strict=True):
            concrete[pos] = sub
        text = "".join(concrete)
        seen.add("true" if text[: len(lhs)] == text[len(lhs) :] else "false")
    return sorted(seen)


def category(v: CtyValue[Any]) -> str:
    v, _ = v.unmark()
    if v.is_unknown:
        return "unknown"
    return "true" if v.value is True else "false"


def render(v: CtyValue[Any]) -> str:
    _, marks = v.unmark()
    text = category(v)
    if v.is_unknown and not getattr(v.value, "is_known_null", None) is False:
        text += " (may be null)"
    if marks:
        text += " marks=" + ",".join(sorted(str(m) for m in marks))
    return text


def run_equals() -> None:
    for kind in ("object", "map", "list", "tuple"):
        for n in range(1, 5):
            for lhs in cell_strings(n):
                for rhs in cell_strings(n):
                    for marked in (False, True):
                        if marked and n > 3:
                            continue
                        left, right = equals_operand(kind, lhs, marked), equals_operand(kind, rhs, False)
                        seen: dict[str, str] = {}
                        for _ in range(EQUALS_ITERATIONS):
                            got = left.equals(right)
                            seen[render(got)] = category(got)
                        emit(
                            {
                                "kind": kind,
                                "lhs": lhs,
                                "rhs": rhs,
                                "marked": marked,
                                "observed": [{"category": c, "gostring": r} for r, c in sorted(seen.items())],
                                "truth": equals_truth(lhs, rhs),
                            }
                        )


# ---------------------------------------------------------------------------
# setproduct
# ---------------------------------------------------------------------------


def collection(length: int) -> CtyValue[Any]:
    return CtyList(element_type=STRING).validate([str(i) for i in range(length)])


def values_sha256(result: CtyValue[Any]) -> str:
    """The same language-neutral digest the Go harness computes: one line per
    tuple, its string elements joined by commas, in result order."""
    h = hashlib.sha256()
    for item in result.value:
        h.update((",".join(str(e) for e in item.raw_value) + "\n").encode())
    return h.hexdigest()


LENGTH_SPECS: list[list[tuple[int, int]]] = [
    *([(n, 2)] for n in range(2, 17)),
    *([(n, 3)] for n in range(2, 9)),
    [(4, 10)],
    [(48, 2)],
    [(56, 2)],
    [(62, 2)],
    [(63, 2)],
    [(64, 2)],
    [(65, 2)],
    [(21, 8)],
    [(22, 8)],
    [(16, 16)],
    [(32, 4)],
    [(8, 256)],
    [(7, 256), (1, 128)],
    [(41, 3)],
    [(64, 2), (1, 0)],
    [(8, 256), (1, 0)],
]


def run_setproduct_lengths() -> None:
    for groups in LENGTH_SPECS:
        expected = 1
        args = []
        for count, length in groups:
            for _ in range(count):
                expected *= length
                args.append(collection(length))
        wrapped = expected % (1 << 64)
        case: dict[str, Any] = {
            "groups": [{"count": c, "length": n} for c, n in groups],
            "expected": str(expected),
            "fits_int": expected <= (1 << 63) - 1,
            "wrapped_int": wrapped - (1 << 64) if wrapped >= 1 << 63 else wrapped,
        }
        try:
            got = setproduct(*args)
        except Exception as exc:
            case.update(outcome="error", length=0, error=described(exc))
        else:
            case.update(outcome="ok", length=len(got.value))
            if case["length"] <= 100_000:
                case["values_sha256"] = values_sha256(got)
        emit(case)


def run_setproduct_memory() -> None:
    got = setproduct(*[collection(10) for _ in range(6)])
    emit({"total_alloc_bytes": [], "length": len(got.value), "values_sha256": values_sha256(got)})


# ---------------------------------------------------------------------------
# Number refinement bounds
# ---------------------------------------------------------------------------

BOUND_VALUES = ["-1", "0", "0.5", "1", "100000000000000000000000000000000000001"]


def refine_bounds(
    lower: str, upper: str, lower_inc: bool, upper_inc: bool, lower_first: bool
) -> tuple[CtyValue[Any] | None, str]:
    try:
        b = refine(CtyValue.unknown(NUMBER))
        lo, hi = Decimal(lower), Decimal(upper)
        if lower_first:
            b = b.number_range_lower_bound(lo, inclusive=lower_inc).number_range_upper_bound(hi, inclusive=upper_inc)
        else:
            b = b.number_range_upper_bound(hi, inclusive=upper_inc).number_range_lower_bound(lo, inclusive=lower_inc)
        return b.new_value(), ""
    except Exception as exc:
        return None, described(exc)


def run_bounds() -> None:
    for lower in BOUND_VALUES:
        for upper in BOUND_VALUES:
            for lower_inc in (False, True):
                for upper_inc in (False, True):
                    for lower_first in (True, False):
                        v, refused = refine_bounds(lower, upper, lower_inc, upper_inc, lower_first)
                        case: dict[str, Any] = {
                            "lower": lower,
                            "upper": upper,
                            "lower_inclusive": lower_inc,
                            "upper_inclusive": upper_inc,
                            "order": "lower-first" if lower_first else "upper-first",
                            "accepted": v is not None,
                        }
                        if v is None:
                            case["panic"] = refused
                        else:
                            case["gostring"] = repr(v.value)
                        emit(case)


def probes(v: CtyValue[Any]) -> list[dict[str, str]]:
    ret = []
    for n in ("2.5", "3", "3.5"):
        num = NUMBER.validate(Decimal(n))
        ret.append({"probe": f"Range().Includes({n})", "result": category(value_range(v).includes(num))})
        ret.append({"probe": f"Equals({n})", "result": category(v.equals(num))})
    return ret


def run_bounds_consequences() -> None:
    v, refused = refine_bounds("3", "3", False, False, True)
    case: dict[str, Any] = {"accepted": v is not None}
    if v is None:
        case["panic"] = refused
    else:
        case["gostring"] = repr(v.value)
        case["consequences"] = probes(v)
    emit(case)


def run_bounds_decode() -> None:
    def payload(lower: int, lower_inc: bool, upper: int, upper_inc: bool) -> str:
        def flag(b: bool) -> str:
            return "c3" if b else "c2"

        body = f"820392{lower:02x}{flag(lower_inc)}0492{upper:02x}{flag(upper_inc)}"
        return f"c7{len(body) // 2:02x}0c{body}"

    cases = [
        ("3 <= x <= 4 (non-empty)", payload(3, True, 4, True)),
        ("3 <= x <= 3 (exactly 3)", payload(3, True, 3, True)),
        ("3 < x < 3 (empty, both exclusive)", payload(3, False, 3, False)),
        ("3 < x <= 3 (empty, lower exclusive)", payload(3, False, 3, True)),
        ("4 <= x <= 3 (empty, lower above upper)", payload(4, True, 3, True)),
    ]
    for desc, hex_payload in cases:
        case: dict[str, Any] = {"description": desc, "payload_hex": hex_payload}
        try:
            v = cty_from_msgpack(bytes.fromhex(hex_payload), NUMBER)
        except Exception as exc:
            case.update(outcome="error", detail=described(exc))
        else:
            case.update(outcome="value", detail=repr(v.value), consequences=probes(v))
        emit(case)


COMMANDS: dict[str, Callable[[], None]] = {
    "equals": run_equals,
    "setproduct-lengths": run_setproduct_lengths,
    "setproduct-memory": run_setproduct_memory,
    "bounds": run_bounds,
    "bounds-consequences": run_bounds_consequences,
    "bounds-decode": run_bounds_decode,
}

if __name__ == "__main__":
    if len(sys.argv) != 2 or sys.argv[1] not in COMMANDS:
        sys.stderr.write(f"usage: harness.py {'|'.join(COMMANDS)}\n")
        sys.exit(2)
    COMMANDS[sys.argv[1]]()
