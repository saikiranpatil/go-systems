# Build Your Own JSON Parser in Go

A self-paced course, built the way boot.dev's HTTP-from-scratch course does it: tiny
verifiable steps, each one a small Go package/function you can test in isolation,
stacking up into a real thing. You already did this once — `01-tcp-echo-server` →
`02-http-server` in your `go-systems` repo is exactly this pattern (raw bytes →
structured protocol). This course is the same muscle, aimed at a different target:
raw bytes → structured **data**, no network required.

Suggested repo location: continue your existing numbering —
`go-systems/03-json-parser/` with packages `lexer/`, `parser/`, `jsonerr/`, and a
`cmd/jsonparse/` CLI, plus a `_testdata/` folder for the JSONTestSuite files.

**How each lesson works:** a one-line goal, what to build, and a "definition of
done" — a test you write yourself (Go's `testing` package, table-driven) that has
to pass before you move on. There's no autograder here like boot.dev's CLI tests,
so *you* are the grader — write the test before or right after the code, not days
later.

**Pace:** 2–4 lessons per sitting is plenty. Unit 1 (lexing) is the biggest one and
where most of the real learning happens — don't rush it.

---

## Unit 0 — Setup & Mental Model

### 0.1 Scaffold and echo
**Goal:** get a project that compiles and does the most boring possible thing.
- `go mod init`, create `lexer/`, `parser/`, `jsonerr/`, `cmd/jsonparse/`.
- `cmd/jsonparse/main.go` takes a hardcoded JSON string, and just prints it back.
- **Done when:** `go run ./cmd/jsonparse` prints your string unchanged.

### 0.2 Reading input three ways
**Goal:** decouple "getting bytes" from "parsing bytes" — you'll need all three later.
- [ ] Read from a Go string literal.
- [ ] Read from `os.Stdin` (so you can `cat file.json | jsonparse`).
- [ ] Read from a file path via `os.Open` + `io.ReadAll`.
- **Done when:** all three paths print the same byte count for the same content.

### 0.3 Read the spec (no code)
**Goal:** know the grammar before you encode it.
- Read [RFC 8259](https://datatracker.ietf.org/doc/html/rfc8259) (short — the
  grammar section is what matters) and skim [json.org](https://www.json.org/).
- Write, in your own words, the 6 JSON value types and the exact rules for a valid
  number and a valid string. You'll need these on recall in Unit 1.
- **Done when:** you can answer without looking: is `01` a valid number? Is
  `{'a':1}` valid JSON? Is a trailing comma in an array valid?

---

## Unit 1 — Lexing: turning bytes into tokens

This is where you build the muscle. Go byte-by-byte, no `regexp`, no shortcuts —
that's the point of the exercise.

### 1.1 Token types
**Goal:** define the vocabulary before writing any logic.
- `TokenType` (LBRACE, RBRACE, LBRACKET, RBRACKET, COLON, COMMA, STRING, NUMBER,
  TRUE, FALSE, NULL, EOF, ILLEGAL) and a `Token{Type, Literal, Pos}` struct.
- No lexing logic yet — this compiles with an empty `Lexer`.

### 1.2 Whitespace and structural tokens
**Goal:** the simplest possible lexer that does something real.
- Walk the input byte by byte. Skip whitespace — **use the JSON definition**
  (space, tab, `\n`, `\r`), not `unicode.IsSpace`, they're not the same set.
- Emit the six single-byte structural tokens: `{ } [ ] : ,`.
- **Done when:** tokenizing `"{ } [ ] , : "` returns exactly those 6 tokens.

### 1.3 Literals: true, false, null
**Goal:** match fixed keyword sequences, reject partial matches.
- Match exact byte sequences; `tru`, `True`, `nulll` are all `ILLEGAL`, not `NULL`.
- **Done when:** a table test covering `true`, `false`, `null`, and 3–4 near-misses
  passes.

### 1.4 Numbers
**Goal:** the first genuinely fiddly grammar rule.
- Build it in layers, testing after each:
  1. Bare integers (`0`, `42`).
  2. Reject leading zeros (`01` is invalid, `0` alone is fine).
  3. Negative sign.
  4. Fraction part (`3.14`).
  5. Exponent part (`1e10`, `1E-5`, `1e+5`).
- **Done when:** you have a table test with ~15 cases (valid and invalid) straight
  out of the spec grammar, e.g. `1.` and `.1` and `1e` should all fail.

### 1.5 Strings
**Goal:** the second fiddly rule, and the one most hand-rolled parsers get wrong.
- Build in layers:
  1. Plain quoted strings, no escapes.
  2. Simple escapes: `\" \\ \/ \b \f \n \r \t`.
  3. Unicode escapes `\uXXXX`.
  4. Surrogate pairs (`\uD83D\uDE00` → a single emoji) — this is the hard part.
  5. Reject raw (unescaped) control characters inside strings.
- **Done when:** your table test includes a surrogate pair and a raw control
  character case, and both behave per spec.

### 1.6 Position tracking
**Goal:** lay groundwork for Unit 3 — you can't point at an error later if you
don't track *where* you are now.
- Track byte offset (and/or line + column) on every emitted token.
- **Done when:** tokenizing a 3-line JSON snippet gives correct line/col for a
  token on line 3.

### 1.7 Full lexer integration
**Goal:** one coherent API, end to end.
- Decide: `Tokenize(input) []Token` (whole-slice) or `NextToken()` (streaming,
  more like Go's own `text/scanner`). Either is fine — streaming pairs better with
  Unit 5's streaming-parser stretch goal if you want to keep that door open.
- **Done when:** a realistic nested JSON snippet (a small config-file-shaped
  object) tokenizes into exactly the token stream you'd expect, asserted in one
  test.

---

## Unit 2 — Parsing: tokens into a Go value

Recursive descent, hand-rolled. (More on why below.)

### 2.1 Parser skeleton
**Goal:** the classic `cur`/`peek` recursive-descent shape.
- `Parser` wraps your token stream, holds current + peek token, has
  `advance()`/`expect(tt)` helpers and a `parseValue()` dispatcher stub.

### 2.2 Parse primitives
**Goal:** the leaves of the tree first.
- `null` → `nil`, `true`/`false` → `bool`, `NUMBER` → `float64`, `STRING` → `string`.
- **Done when:** each of the 4 primitive kinds round-trips through `parseValue`
  in isolation.

### 2.3 Parse arrays
**Goal:** first recursive rule.
- `[` value (`,` value)* `]`, plus the empty array `[]`.
- **Done when:** `[1, "a", true, null, [1,2]]` parses to the right
  `[]interface{}`, including the nested array.

### 2.4 Parse objects
**Goal:** the other recursive rule, and a design decision.
- `{` string `:` value (`,` string `:` value)* `}`, plus `{}`.
- Decide what happens on a **duplicate key** — the spec leaves this
  implementation-defined. Pick "last one wins" (matches `encoding/json`) and write
  it down in a comment.
- **Done when:** a nested object-of-arrays-of-objects parses correctly, and your
  duplicate-key test does what you decided.

### 2.5 Deep nesting, realistic input
**Goal:** stress-test the recursion with something real.
- Parse a genuine small JSON file (e.g. a `package.json` or a GitHub API sample).
- **Done when:** you can walk the resulting `map[string]interface{}` /
  `[]interface{}` tree and pull out a deeply nested value correctly.

### 2.6 Top-level rules
**Goal:** the easy-to-forget edge of the grammar.
- A JSON text is *exactly one* value — reject empty input, reject trailing
  garbage after a complete value (`{}{}`, `1 2`).
- **Done when:** these are explicit failing test cases, not just "seems to work."

---

## Unit 3 — Errors that don't suck

### 3.1 A real error type
**Goal:** stop using bare `errors.New` everywhere.
- `ParseError{Message, Offset, Line, Col}` implementing the `error` interface.
  Replace every ad-hoc error in the lexer/parser with it.

### 3.2 Pretty printing
**Goal:** point at the mistake, don't just describe it.
- Given the original input + an offset, print the offending line with a `^`
  underneath the bad byte — compiler-style.
- **Done when:** feeding `{"a": tru}` prints something you'd actually want to read.

### 3.3 Context-aware messages
**Goal:** "invalid JSON" is a useless error message.
- At each parse point, say what was expected: `"expected ',' or '}' but found EOF"`
  rather than a generic failure.
- **Done when:** 5 different malformed inputs each produce a distinct, accurate
  message.

---

## Unit 4 — Prove it: JSONTestSuite

### 4.1 Wire up the suite
**Goal:** stop trusting your own test cases, use the industry-standard adversarial
set.
- Vendor `test_parsing/` from
  [github.com/nst/JSONTestSuite](https://github.com/nst/JSONTestSuite) into
  `_testdata/`.
- Write one Go test that walks the directory: `y_*.json` must parse successfully,
  `n_*.json` must fail, `i_*.json` (implementation-defined) just gets logged, not
  asserted.

### 4.2 Triage failures
**Goal:** this is where Unit 1's corner cases actually get stress-tested for real.
- Go failure by failure. Each one maps back to a specific lexer/parser lesson
  above — fix it there, not with a patch bolted onto Unit 4.
- **Done when:** all `y_` and `n_` cases pass; you've made and documented a
  deliberate call on each `i_` case you hit.

### 4.3 Compare against `encoding/json`
**Goal:** catch *semantic* bugs, not just accept/reject bugs.
- For every `y_` file, unmarshal with both `encoding/json` and your parser, and
  deep-compare the resulting values.
- **Done when:** the two agree on every case in the suite.

---

## Unit 5 — Stretch goals (optional, "adventurous" tier)

Pick any of these once Unit 4 is green — they're independent of each other.

- **5.1 Streaming API:** expose a pull-based `NextToken()`/`More()` interface like
  `encoding/json.Decoder`, so huge files don't need to fit in memory as one tree.
- **5.2 Marshal back to JSON:** write `Marshal(v interface{}) ([]byte, error)` so
  `parse → marshal → parse` round-trips to the same value.
- **5.3 Benchmark it:** `go test -bench` against `encoding/json` on a real-world
  payload (e.g. the `twitter.json` / `canada.json` files used in most JSON
  parser benchmarks). This is where "why is JSON parsing slow" stops being a
  trivia question and becomes something you measured yourself.
- **5.4 Try goyacc, just to know why you didn't:** generate a parser for the same
  grammar with `goyacc` and compare it to your hand-rolled version. For a grammar
  this small, most people end up preferring recursive descent — worth confirming
  that for yourself rather than taking it on faith.

---

## A note on skipping Lex/Yacc

The original brief you shared frames Lex/Yacc as the default and hand-rolled as
the "adventurous" option. I've flipped that: Unit 2 above is hand-rolled
recursive descent by default. Two reasons — it's the idiomatic Go approach (it's
literally how `go/parser` and `encoding/json` are built internally), and for a
grammar this small, generator tooling adds ceremony without adding insight. 5.4
is there if you want to confirm that trade-off for yourself rather than take my
word for it.

## Reference links
- [RFC 8259 — The JSON Data Interchange Format](https://datatracker.ietf.org/doc/html/rfc8259)
- [json.org grammar diagrams](https://www.json.org/json-en.html)
- [nst/JSONTestSuite](https://github.com/nst/JSONTestSuite)
- *Writing an Interpreter in Go* (Thorsten Ball) — not JSON, but the exact
  cur/peek recursive-descent shape used in Unit 2, if you want a second worked
  example of the pattern.