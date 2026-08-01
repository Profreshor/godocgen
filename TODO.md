# godocgen — Roadmap

Ordered by impact. Cheap correctness wins first, structural changes next,
new languages after the seam is proven, cosmetics last.

Sizes: **S** = under an hour. **M** = an evening. **L** = a weekend or more.

Every finding below was reproduced against the binary. Evidence is included so
this file stands on its own.

---

## Phase 0 — Safety net

Nothing else on this list is safe without this. Phases 2 and 3 rewrite the
lexer core. You cannot verify those changes by reading the diff.

- [ ] **Lexer table tests** — S
  Input: a source string. Assert: token kinds and spans.
  Cover: idents, numbers, all three string forms, `//` and `/* */`, operator
  greediness (`<<=` must beat `<<` must beat `<`), EOF emission.
  *Done when:* changing a `Literals` entry breaks exactly one test.

- [ ] **Parser table tests** — S
  `CreateParser` takes `([]lexer.Token, []byte)` — pure data, no mocks needed.
  Cover: func, method, grouped `var`/`const`/`type`, grouped imports, aliased
  imports, generics, doc comment attachment.
  *Done when:* every bug in Phase 1 has a failing test before you fix it.

- [ ] **Assemble tests** — S
  The `testdata/evil` fixtures already target the hard cases. Use them.
  *Done when:* orphan detection and the var/type name collision are both pinned.

---

## Phase 1 — Bugs that silently corrupt output

All three are small. All three are wrong *today*, on your own repos.

- [ ] **`collectDoc` drops the doc comment on the last declaration in a file** — S
  `parser.go:306` guards a backwards loop over `i` with `p.isValid()`, which
  tests `p.pos`. After `scanToLineEnd` on the final declaration, `p.pos` sits on
  EOF, so the loop never starts.

  ```
  // LastVar is the final declaration.     // LastVar doc SHOULD appear now.
  var LastVar = 1                          var LastVar = 1
                                           func Trailing() {}
  → doc DROPPED                            → doc APPEARS
  ```

  *Done when:* the doc survives with nothing after it. Ask what that loop is
  stepping over, and whether `p.pos` has any bearing on it.

- [ ] **`Assemble` output is not reproducible** — S
  `assemble.go:38` ranges over the `methodsByOwner` map. Go randomises map
  order. Eight runs over four orphans gave two different orderings:

  ```
  M1(Alpha) M2(Beta) M3(Gamma) M4(Delta)
  M4(Delta) M1(Alpha) M2(Beta) M3(Gamma)
  ```

  You already solved this for directories in `main.go:69-73` — collect keys,
  sort, iterate. This blocks `ETag`/`Last-Modified` later.
  *Done when:* 20 consecutive runs are byte-identical.

- [ ] **Lexer reads past the end of the source buffer** — S
  `lexer.go:102` does `pos += 2` on an escape with no bounds check. A file
  ending in `\` pushes `pos` to `len+1`.

  It does not panic today only because `os.ReadFile` returns a buffer with
  `cap == len+1`, and Go bounds-checks slice *high* indices against capacity.
  The leak still reaches output — `var S = "abc\` emitted a NUL byte into the
  HTML:

  ```
  r   S   =   &#34;   a   b   c   \  \0
                                     ^^ past end of file
  ```

  If `Content` ever comes from `[]byte(str)`, a sub-slice, or `io.ReadAll`,
  `cap == len` and this becomes a hard panic — attacker-controlled once served.
  *Done when:* `pos` cannot exceed `len(source)`. Prefer one choke point over
  a check at each call site.

---

## Phase 2 — The change that gets more expensive every week

- [ ] **Rune-based lexing (UTF-8)** — M
  `tokens.go:72` accepts ASCII and `_` only. Go allows any Unicode letter in an
  identifier. The symbol does not appear with a wrong name — it vanishes:

  ```
  func Ünicode() {}   → no symbol at all
  func Ascii()    {}  → Function Ascii
  ```

  Markdown is prose. Prose is full of non-ASCII. Do this before Markdown, not
  after. Look at `utf8.DecodeRune` and `unicode.IsLetter`.

  Decide deliberately: **spans stay byte offsets** (you slice `source` with
  them), but advance by rune width.

- [ ] **Add line and column to `Pos` — while you are already in there** — S
  `tokens.go:3` wraps a single `int`. A bare struct around one field exists to
  be extended. Without line numbers the served docs cannot link a symbol to its
  source line, which is the main thing a reader wants.

  Fold this into the rune refactor. You are rewriting every `pos++` anyway, and
  the lexer is the only place that sees every newline cheaply. Doing it as a
  separate pass means touching the same code twice.

---

## Phase 3 — Make the lexer data-driven

The enabling refactor. No new language yet. Three things are hardcoded in the
core that must become `Language` table data:

- [ ] **Comment syntax → `Language`** — M
  `//` and `/*` are baked into `consumePunctOrOper` (`lexer.go:126`, `:133`).
  Python needs `#`. Markdown needs `<!-- -->`.

- [ ] **String delimiters → `Language`** — M
  `` ` ``, `"`, `'` are literal cases in the `Tokenize` switch
  (`lexer.go:41-46`). Python needs `"""` and prefixes (`r"`, `f"`, `b"`).

- [ ] **Walker takes `fs.FS` instead of a path string** — S
  `walker.go:28-30` builds `os.DirFS` *inside* the function, sealing the
  abstraction in. Accept the `fs.FS` instead and tests can use `fstest.MapFS` —
  in-memory, no disk, no fixtures on disk.

  Catch: `walker.go:46` calls `os.ReadFile(absPath)`, bypassing the filesystem
  value. Both must change together.

  This is the interface exercise. You design nothing — you notice you already
  fit one the standard library wrote.

---

## Phase 4 — Markdown (second, not last)

Counterintuitive, but do it here **because it does not fit**. It forces the
seam while the codebase is still ~1400 lines. Do Python first and you will bend
`Language` to fit, it will mostly work, and Markdown will then demand a rewrite
of a larger system.

- [ ] **Write the Markdown extractor as a plain concrete function** — M
  No interface. Nothing to satisfy. Markdown has no identifiers, keywords,
  operators or braces — it is line-oriented. Do not force it through
  `lexer.go`. Probably ~80 lines of line scanning.

- [ ] **Only now: compare the two and extract the interface** — S
  Put the Go path and the Markdown path side by side. Three questions you
  cannot answer today:
  - Does an extractor need `path`, or only `src`?
  - Does it return an `error`? Go's path never fails — it skips. If only one
    language needs an error, that says something about the whole design.
  - `[]byte` or `io.Reader`? Your lexer indexes randomly into `source` to build
    spans, so streaming is impossible. That constraint is invisible from
    outside.

  Sketch only:
  ```
  Extract(src, path) -> ([]Symbol, []Note)
  registry: ".go" -> goExtractor, ".md" -> mdExtractor
  ```
  Put it where it is *consumed*. No extractor should import it to satisfy it.

- [ ] **Decide what `Symbol` means across languages** — M
  Does `SymbolKind` grow per language (HEADING, CLASS, DECORATOR), or map onto
  a shared vocabulary? You left commented-out kinds at `symbols.go:22-44`, so
  you have been circling this. Which choice keeps `render` from growing a
  switch per language?

  Also: "package" is a directory in Go, a file in Python, and nothing in
  Markdown. `main.go:66` groups by directory — a Go assumption. What is the
  grouping unit in a mixed repo?

---

## Phase 5 — HTTP

- [ ] **Decide: index once at startup, or parse per request** — S
  Answer this before writing a handler. Everything follows from it. Per-request
  is simple and always fresh but re-lexes the repo on every refresh. Index-once
  is fast and cacheable but you own invalidation and shared state.

  If you index once: `assemble.go:32` does `topLevel[i].Children = append(...)`
  — mutation in place. What does that do to a shared index read by many
  goroutines?

- [ ] **Buffer the render before choosing a status code** — S
  I praised `pageWriter.firstError` in review. Over HTTP it inverts. The first
  `Write` to a `ResponseWriter` commits `200 OK` and flushes headers, so when
  `Render` returns an error at `html.go:62` the client already has half a page
  and a success status.

  Render into memory, check the error, *then* pick the status and copy out.
  Bonus: you get `Content-Length` and a cheap `ETag`.

- [ ] **Escape `anchor()` — or move to `html/template`** — S
  `anchor()` (`html.go:80`) never escapes, and its output is injected raw into
  two attributes (`html.go:49-51`, `html.go:66`). A directory name with a quote
  in it — legal on Linux — breaks out:

  ```
  <a href="#pkg-evil"onmouseover=alert(1) x">
  <section id="pkg-evil"onmouseover=alert(1) x">
  ```

  The link *text* is escaped correctly; the attribute is not. Today it needs a
  hostile directory name on your own disk. Served over HTTP against repos you
  did not write, it is stored XSS.

  Prefer `html/template`: it is context-aware — it knows an attribute from a
  text node from a URL — which is exactly why `html.EscapeString` missed this.

- [ ] **Never map a URL to a filesystem path** — S
  Key into your index by an ID you generated. If you must touch the filesystem,
  read up on `filepath.Clean` semantics and `os.Root` (Go 1.24+).

- [ ] **Retire the debug output** — S
  Delete `internal/report` (zero callers), `printSymbol` (`main.go:105`), and
  the `output/out.html` sink. Two reporting implementations exist and the
  better-looking one never runs.

- [ ] **Lift the pipeline out of `main`** — S
  `main.go:47-78` is walk, lex, parse, group, sort, assemble, mkdir, render —
  eight responsibilities. `main` should parse arguments and report errors.
  The server will need this as a callable function anyway.

---

## Phase 6 — Python

- [ ] **INDENT / DEDENT tokens** — L
  `Tokenize` discards whitespace (`lexer.go:34-36`). Python cannot. This means
  a stack of indent widths and synthesised tokens that correspond to **no bytes
  in the source**. What is the `Span` of a DEDENT? Decide before you write it.

- [ ] **`Owner` is a Go-ism — confront it here** — M
  Go declares methods at top level with a receiver, so you resolve them to a
  type *by name* afterwards. That is the only reason `Assemble` and the
  "orphaned method" concept exist.

  Python and JS/TS nest methods lexically inside the class body. The parent is
  known structurally at parse time. Orphans are impossible.

  So `assemble.go` is not a shared pipeline stage — it is the Go backend's
  fixup pass. What happens to it when three of four languages skip it?

---

## Phase 7 — JS/TS

- [ ] **Template literals** — L
  `` `a${expr}b` `` requires the lexer to recurse into itself mid-string.
- [ ] **JSX** — L
  A different grammar appearing mid-expression.
- [ ] **TypeScript types** — L
  Effectively a second language layered on the first.

You were right that this is the painful one. Do not start it until the
`Extractor` seam has survived two dissimilar languages.

---

## Sugar

Cosmetics and hygiene. None of it blocks anything.

- [ ] `git rm --cached .DS_Store bin/godocgen` — S
  Both are tracked. `.gitignore` lists `/bin/`, but the rule cannot affect an
  already-committed file. A 3 MB binary in history will only grow.
- [ ] Add `.DS_Store` to `.gitignore` — S
- [ ] `.note` class is emitted (`html.go:74`) but has no CSS rule — S
- [ ] `anchor()` can collide: it maps `/` to `-`, so `a/b` and `a-b` produce the
  same DOM id — S
- [ ] Parenthesise the condition at `lexer.go:90` — S
  `a && b || c` parses as `(a && b) || c`. The `isValid()` guard does not cover
  the `peek() == '.'` test. It happens not to loop forever because `peek`
  returns `0` past the end, but the code reads as if the guard covers both.
- [ ] `Assemble` and `AssemblePackage` disagree on the package name — S
  First MODULE wins at `assemble.go:20`, last wins at `:72`. On `twin_a` the
  tree says `twina` and the heading says `twin_a`. Only the invalid fixture
  triggers it, but the rules genuinely conflict.
- [ ] `gofmt -l .` fails on the intentional `twin_a` fixture — S
  `go vet` and `go build` correctly ignore `testdata/`; `gofmt` walks it.
  Worth knowing before you add CI.
- [ ] Unify per-file error handling — S
  A lexer error calls `os.Exit(1)` and kills the run (`main.go:56`); a load
  error skips the file (`main.go:51`). Both are per-file problems.

---

## The two rules worth keeping

**Do the things that get harder later, first.** Rune handling and line/column
tracking touch every advance in the lexer. That code triples in size once three
more languages live in it.

**Write the second implementation before the interface that covers both.** An
interface with one implementation is a hypothesis. With two, it is an
observation. Your stage boundaries are plain data — tokens in, symbols out,
`PackageDoc` to the renderer — and data is the most flexible contract there is.
Reach for an interface only when one *side* of a boundary needs to vary. Right
now exactly one does, and only once Markdown lands.
