 godocgen — Roadmap

This document gives the work in order of importance. Correct the cheap defects
first. Then make the structural changes. Then add the new languages. Do the
small items last.

Sizes: **S** = less than one hour. **M** = one evening. **L** = one weekend or
more.

The author found each defect below with the compiled program. This document
includes the evidence. You do not need the chat record.

This document keeps the technical names of the project. Examples are `lexer`,
`parser`, `token`, `slice`, and `span`.

---

## Phase 0 — Make the tests

The other phases change the core of the `lexer`. You cannot check these changes
when you read the difference in the code. Make the tests first.

- [ ] **Make the table tests for the `lexer`** — S
  Give a source text. Check the token kinds and the spans.
  Test these items:
  - identifiers
  - numbers
  - the three forms of string
  - the `//` comment and the `/* */` comment
  - the longest operator match (`<<=` before `<<` before `<`)
  - the EOF token

  *Correct when:* a change to one `Literals` entry causes one test to fail.

- [ ] **Make the table tests for the `parser`** — S
  `CreateParser` receives `([]lexer.Token, []byte)`. This is data only. You do
  not need a mock object.
  Test these items:
  - a function and a method
  - a group of `var`, `const`, and `type`
  - a group of imports
  - an import with an alias
  - a generic function
  - the attachment of a doc comment

  *Correct when:* each defect in Phase 1 has a test that fails before you
  correct the defect.

- [ ] **Make the tests for `Assemble`** — S
  The fixtures in `testdata/evil` show the difficult conditions. Use them.
  *Correct when:* the tests hold the orphan method behavior and the collision
  of the `var` name and the `type` name.

---

## Phase 1 — Defects that make the output incorrect

Each defect is small. Each defect is incorrect now, in your own repositories.

- [ ] **`collectDoc` removes the doc comment from the last declaration in a
  file** — S

  The loop at `parser.go:306` moves backwards through `i`. But the condition
  uses `p.isValid()`, which examines `p.pos`. After `scanToLineEnd` operates on
  the last declaration, `p.pos` is at the EOF token. Thus the loop does not
  start.

  ```
  // LastVar is the final declaration.     // LastVar doc SHOULD appear now.
  var LastVar = 1                          var LastVar = 1
                                           func Trailing() {}
  → doc REMOVED                            → doc CORRECT
  ```

  *Correct when:* the doc stays when no declaration comes after it.

  Question: what does the loop examine? Does `p.pos` have an effect on it?

- [ ] **`Assemble` gives a different result at each run** — S
  The loop at `assemble.go:38` moves through the `methodsByOwner` map. Go makes
  the map order random. Eight runs on four orphan methods gave two orders:

  ```
  M1(Alpha) M2(Beta) M3(Gamma) M4(Delta)
  M4(Delta) M1(Alpha) M2(Beta) M3(Gamma)
  ```

  You corrected this condition for the directories at `main.go:69-73`. You
  collect the keys, you sort them, then you move through them.

  This defect prevents `ETag` and `Last-Modified` in Phase 5.
  *Correct when:* 20 runs give the same bytes.

- [ ] **The `lexer` reads bytes after the end of the source** — S
  The code at `lexer.go:102` adds 2 to `pos` for an escape. It does not examine
  the limit. A file that ends with `\` moves `pos` to `len+1`.

  **The condition for the panic is `pos > cap(source)`.** The condition is not
  `cap == len`. A slice expression examines the high index against the
  capacity, not against the length.

  Do not rely on the capacity. The capacity changes with the Go version, the
  memory size class, and the escape analysis. These are the measured values for
  a source of 25 bytes:

  ```
  os.ReadFile      len=25 cap=512   no panic
  io.ReadAll       len=25 cap=512   no panic
  sub-slice        len=25 cap=90    no panic
  []byte(str)      len=25 cap=25    PANIC
  b[:n:n]          len=25 cap=25    PANIC
  ```

  The program does not stop today because `os.ReadFile` gives a large capacity.
  But the incorrect byte goes into the output. The source `var S = "abc\` put a
  NUL byte into the HTML:

  ```
  r   S   =   &#34;   a   b   c   \  \0
                                     ^^ byte after the end of the file
  ```

  For a test that always causes the panic, use a full slice expression:
  `b[:n:n]` makes the capacity equal to the length. The measured result is
  `slice bounds out of range [:26] with capacity 25`.

  An attacker can cause this panic when you supply the program with HTTP.

  *Correct when:* `pos` cannot become more than `len(source)`.

  Question: is it better to examine the limit at one control point, or at each
  location that moves `pos`?

---

## Phase 2 — The change that becomes more difficult each week

- [ ] **Change the `lexer` to operate on runes (UTF-8)** — M
  The function at `tokens.go:72` accepts ASCII letters and `_` only. Go permits
  all Unicode letters in an identifier. The symbol does not get an incorrect
  name. The symbol does not appear:

  ```
  func Ünicode() {}   → no symbol
  func Ascii()    {}  → Function Ascii
  ```

  Markdown contains prose. Prose contains many non-ASCII characters. Make this
  change before you add Markdown.

  Examine `utf8.DecodeRune` and `unicode.IsLetter`.

  Make this decision first: the spans stay as byte offsets, because you use
  them to cut the `source`. But `pos` must move by the width of the rune.

- [ ] **Add the line number and the column number to `Pos`** — S
  The structure at `tokens.go:3` contains one `int`. A structure that contains
  one field is a structure that you will extend.

  Without a line number, the HTML pages cannot make a link to the source line.
  This link is the primary function of a documentation tool.

  Make this change during the rune change. You change each movement of `pos` in
  that task. The `lexer` is the only component that sees each newline at a low
  cost. If you do this task later, you must change the same code two times.

---

## Phase 3 — Move the language rules into data

This phase adds no new language. Three rules are in the core of the `lexer`.
They must become data in the `Language` structure.

- [ ] **Move the comment syntax into `Language`** — M
  The `//` mark and the `/*` mark are in `consumePunctOrOper` at `lexer.go:126`
  and `lexer.go:133`. Python uses `#`. Markdown uses `<!-- -->`.

- [ ] **Move the string delimiters into `Language`** — M
  The `` ` `` mark, the `"` mark, and the `'` mark are in the `Tokenize` switch
  at `lexer.go:41-46`. Python needs `"""`. Python also needs the prefixes `r"`,
  `f"`, and `b"`.

- [ ] **Give an `fs.FS` value to the walker** — S
  The function at `walker.go:28-30` makes the `os.DirFS` value inside itself.
  Thus the caller cannot supply a different file system.

  Receive the `fs.FS` value as a parameter. Then the tests can use
  `fstest.MapFS`. This file system is in memory. It needs no disk and no
  fixture files.

  Caution: the code at `walker.go:46` calls `os.ReadFile(absPath)`. This call
  does not use the file system value. You must change both parts together.

  This task is the interface exercise. You design no interface. You see that
  your code agrees with an interface from the standard library.

---

## Phase 4 — Markdown (the second language, not the last)

Add Markdown here because Markdown does not agree with the `lexer` design. It
shows you the correct interface boundary while the code is still near 1400
lines.

If you add Python first, you will change the `Language` structure to fit
Python. The result will operate correctly for Python. But the structure will
then be incorrect for Markdown, and the code will be larger.

- [ ] **Write the Markdown extractor as a usual function** — M
  Write no interface. Markdown has no identifiers, no keywords, no operators,
  and no braces. Markdown has lines. Do not send Markdown through `lexer.go`.
  The result is near 80 lines that examine lines.

- [ ] **Compare the two extractors. Then make the interface** — S
  Put the Go extractor and the Markdown extractor together. Answer these
  questions. You cannot answer them today:

  - Does an extractor need the `path`, or the `src` only?
  - Does an extractor return an `error`? The Go extractor cannot fail. It
    ignores the parts that it cannot read. If one language only needs an
    `error`, what does this tell you about the design?
  - Does an extractor receive a `[]byte` or an `io.Reader`? Your `lexer` moves
    to random positions in the `source` to make the spans. Thus a stream is not
    possible. You cannot see this limit from outside the function.

  Pseudocode only:
  ```
  Extract(src, path) -> ([]Symbol, []Note)
  registry: ".go" -> goExtractor, ".md" -> mdExtractor
  ```

  Put the interface in the package that uses it. No extractor imports the
  interface to agree with it.

- [ ] **Decide what a `Symbol` is in each language** — M
  Does `SymbolKind` get more values for each language (HEADING, CLASS,
  DECORATOR)? Or do you map each language onto one set of values? You made
  comments for more kinds at `symbols.go:22-44`. Thus you thought about this
  question.

  Which decision prevents a switch for each language in `render`?

- [ ] **Decide the group unit for a repository with more than one language** — M
  The word "package" has a different meaning in each language:

  - **Go:** a package is a directory.
  - **Python:** a module is a `.py` file. A regular package is a directory that
    contains an `__init__.py` file. A namespace package (PEP 420) contains no
    `__init__.py` file, and it can go across more than one directory.
  - **Markdown:** there is no package.

  The code at `main.go:66` makes the groups by directory. This is correct for
  Go only. The Python namespace package breaks the rule of one package for one
  directory.

  Question: what is the group unit when a repository contains all four
  languages?

---

## Phase 5 — The HTTP server

- [ ] **Decide: make the index one time, or read the files at each request** — S
  Answer this question before you write a handler. The other decisions come
  from it.

  A read at each request is simple, and the data is always new. But the program
  reads the full repository at each refresh of the page.

  An index that you make one time is fast, and you can cache it. But you must
  then remove the old data at the correct time, and more than one goroutine
  reads the same data.

  If you make the index one time, examine `assemble.go:32`. The code
  `topLevel[i].Children = append(...)` changes the data in position. What is
  the result when many goroutines read this index?

- [ ] **Put the HTML in memory before you select the status code** — S
  The review gave a good opinion of the `firstError` field in `pageWriter`. For
  HTTP, this design is a problem.

  The first `Write` to a `ResponseWriter` sends the status `200 OK` and the
  headers. Thus, when `Render` returns an error at `html.go:62`, the client
  already has one half of the page and a status that shows success.

  Write the HTML into memory. Examine the error. Then select the status code
  and send the bytes.

  This design also gives you the `Content-Length` header and a low-cost `ETag`.

- [ ] **Escape the output of `anchor()`, or use `html/template`** — S
  The function `anchor()` at `html.go:80` does not escape its result. The code
  puts this result into two HTML attributes at `html.go:49-51` and
  `html.go:66`. A directory name that contains a quotation mark is permitted on
  Linux. Such a name goes out of the attribute:

  ```
  <a href="#pkg-evil"onmouseover=alert(1) x">
  <section id="pkg-evil"onmouseover=alert(1) x">
  ```

  The text of the link is correct. The attribute is not correct.

  Today, an attacker needs a directory with a special name on your disk. With
  HTTP and a repository from a different author, this defect is stored XSS.

  Use `html/template`. It knows the difference between an attribute, a text
  node, and a URL. The function `html.EscapeString` does not know the position
  of its result. This is the cause of the defect.

- [ ] **Do not make a file path from a URL** — S
  Use an identifier that you made, and find the data in your index. If you must
  read the disk, examine the behavior of `filepath.Clean` and `os.Root`
  (Go 1.24 and later).

- [ ] **Remove the debug output** — S
  Remove `internal/report`, which has no caller. Remove `printSymbol` at
  `main.go:105`. Remove the `output/out.html` file sink. You have two report
  functions, and the program does not call the better one.

- [ ] **Move the pipeline out of `main`** — S
  The code at `main.go:47-78` does eight tasks: walk, lex, parse, group, sort,
  assemble, make a directory, and render. The `main` function must read the
  arguments and show the errors only.

  The server needs this pipeline as a function.

---

## Phase 6 — Python

- [ ] **Add the INDENT token and the DEDENT token** — L
  The `Tokenize` function removes the whitespace at `lexer.go:34-36`. Python
  cannot remove it. You need a stack of indent widths. You must also make
  tokens that have **no bytes in the source**.

  Question: what is the `Span` of a DEDENT token? Decide before you write the
  code.

- [ ] **Correct the `Owner` design** — M
  Go declares a method at the top level, with a receiver. Thus you must find
  the type by its name after the parse. This is the only reason for `Assemble`
  and for the orphan method.

  Python, JavaScript, and TypeScript put a method inside the class body. The
  parent is known during the parse. An orphan method is not possible.

  Thus `assemble.go` is not a stage of the shared pipeline. It is a correction
  pass for Go.

  Question: what happens to `assemble.go` when three of the four languages do
  not use it?

---

## Phase 7 — JavaScript and TypeScript

- [ ] **Template literals** — L
  The form `` `a${expr}b` `` needs the `lexer` to call itself inside a string.
- [ ] **JSX** — L
  JSX is a different grammar inside an expression.
- [ ] **TypeScript types** — L
  The type syntax is a second language above the first language.

Do not start this phase until two different languages have used the `Extractor`
interface.

---

## Small items

These items are not urgent. They prevent no other task.

- [ ] Run `git rm --cached .DS_Store bin/godocgen` — S
  Git holds both files. The `.gitignore` file has the rule `/bin/`. But a rule
  has no effect on a file that Git holds already. A binary file of 3 MB in the
  history will become larger.
- [ ] Add `.DS_Store` to `.gitignore` — S
- [ ] The `note` class at `html.go:74` has no CSS rule — S
- [ ] The `anchor()` function can make two equal identifiers — S
  It changes `/` to `-`. Thus `a/b` and `a-b` give the same DOM identifier.
- [ ] Add parentheses to the condition at `lexer.go:90` — S
  Go reads `a && b || c` as `(a && b) || c`. Thus the `isValid()` guard has no
  effect on the `peek() == '.'` test. The loop does not continue without an
  end, because `peek` returns `0` after the end of the source. But the code
  looks incorrect.
- [ ] `Assemble` and `AssemblePackage` select a different package name — S
  The first MODULE wins at `assemble.go:20`. The last MODULE wins at
  `assemble.go:72`. For `twin_a`, the tree shows `twina` and the heading shows
  `twin_a`. Only the incorrect fixture causes this result. But the two rules do
  not agree.
- [ ] The command `gofmt -l .` fails on the `twin_a` fixture — S
  The commands `go vet` and `go build` ignore the `testdata` directory. The
  command `gofmt` reads it. Know this before you add CI.
- [ ] Make the error behavior the same for each file — S
  A `lexer` error calls `os.Exit(1)` and stops the program at `main.go:56`. A
  load error ignores the file at `main.go:51`. Both errors apply to one file.

---

## The two primary rules

**Do the difficult tasks first.** The rune change and the line number change
touch each movement of `pos` in the `lexer`. That code will become three times
larger when it holds three more languages.

**Write the second implementation before you write the interface.** An
interface with one implementation is a guess. An interface with two
implementations is a measurement.

Your stage limits are data: tokens in, symbols out, and a `PackageDoc` to the
renderer. Data is the most flexible contract. Use an interface only when one
side of a limit must change. Today, one side must change, and only when you add
Markdown.


### Conversation Context

###### What you did right
The architecture is the strongest part. You split the pipeline into walker → lexer → parser → assemble → render. Each stage has one job. Each stage takes data and returns data. No stage reaches back into another. This is the correct shape for a compiler-style tool, and many people get it wrong on a first attempt.

*You kept the lexer language-agnostic*. Language holds a literal table, and SupportedLanguages maps an extension to a table (languages.go:3). The lexer core never names Go. Adding a second language is a data change, not a logic change. The commented-out entries in walker.go:21 show you already planned this.

*Token-based brace matching is a real insight*. scanTo compares token text, not raw bytes (parser.go:85). So a } inside a string or a comment cannot confuse skipBalanced. Many hand-written parsers get this wrong and need escape logic later. You avoided the problem by construction.

*Error handling in the walker is well judged*. You store LoadErr on the file (walker.go:51) instead of aborting the walk. One unreadable file does not kill the run. That is the right call for a documentation tool.

*pageWriter sticky-error pattern is idiomatic*. You record firstError and make every write a no-op after a failure (html.go:90). This removes an if err != nil from every write call. It matches how bufio.Writer behaves.

*testdata/ naming is exactly right*. Go tooling ignores directories named testdata. Your intentionally-invalid twin_a/config.go therefore does not break go build ./... or go vet ./.... I confirmed both pass clean. That was a deliberate, informed choice.

*Your adversarial fixtures show good instincts*. evil/orphan.go and evil/collision.go target the exact cases that break naive assemblers: a method with no type, and a var sharing a name with a type. You thought about failure before it happened.

*Generics parse correctly*. I tested `func Map[T any, U any](in []T, f func(T) U) []U` and it came through intact. The `[` skip in `readReceiver` (`parser.go:240`) and parseOneSpec (parser.go:165) handle type parameters properly.

###### Confirmed bugs
1. The last declaration in a file silently loses its doc comment. This is the most damaging one, because it loses user data quietly.

collectDoc guards its loop with p.isValid() (parser.go:306). But the loop walks backwards over i. It does not use p.pos. When you parse the final declaration, scanToLineEnd leaves p.pos on the EOF token. isValid() then returns false, and the loop never starts.

Evidence — the same var declaration, with and without a declaration after it:

```Go
// LastVar is the final declaration in this file.
var LastVar = 1                    → doc DROPPED
// LastVar doc SHOULD appear now.
var LastVar = 1
func Trailing() {}                 → doc APPEARS
```

Ask yourself: what is that loop actually stepping over? Does the parser's current position have any bearing on whether token i is a comment?


2. *The lexer reads past the end of the source buffer.* In consumeDelimited, an escape does `lex.pos += 2` with no bounds check (`lexer.go:102`). If a `\` is the last byte of the file, pos becomes len(source)+1. The token span then points beyond the file.

This does not crash today, and the reason is luck: os.ReadFile returns a buffer with capacity len+1, and Go bounds-checks slice high indices against capacity, not length. So sliceSpan reads one byte of allocator slack instead of panicking.

I confirmed the leak reaches your output. A file ending in var S = "abc\ produced this in out.html:

```Go
r   S   =   &   #   3   4   ;   a   b   c   \  \0
                                                ^^ NUL byte from past the 
end
```
The danger is that the safety is accidental. If Content ever arrives from []byte(str), a sub-slice, or io.ReadAll, capacity may equal length, and this becomes a hard slice bounds out of range panic. What invariant should hold every time you advance pos? Where is the single best place to enforce it?

3. Output is not reproducible between runs. Assemble phase 3 ranges over the methodsByOwner map (assemble.go:38). Go randomises map iteration order deliberately. Both the notes and the appended orphan symbols inherit that randomness.

Eight runs over four orphaned methods produced two different orderings:
```
M1(Alpha) M2(Beta) M3(Gamma) M4(Delta)
M4(Delta) M1(Alpha) M2(Beta) M3(Gamma)
```
Generated documentation usually gets committed or diffed. Non-deterministic output makes that painful. You already solved this problem correctly for directories in main.go:69-73 — collect keys, sort, then iterate. Why does the same discipline not apply here?

4. Unicode identifiers are dropped completely. isLetter accepts ASCII and _ only (tokens.go:72). Go permits any Unicode letter in an identifier.
```
func Ünicode() {}   → no symbol at all
func Ascii()    {}  → Function Ascii
```
The function does not appear with a wrong name. It vanishes, because expect(lexer.IDENT) fails on the multi-byte prefix. Your lexer steps one byte at a time, but Go source is UTF-8. What would change if you stepped one rune at a time? Look at utf8.DecodeRune and unicode.IsLetter. Note this also affects Pos.Byte semantics, so decide deliberately whether spans stay byte offsets.

###### Improvements worth making
Tests are the biggest gap. There are zero _test.go files. You have a lexer, a parser, and an assembler — three components with pure functions, table-driven inputs, and comparable outputs. This is the easiest code in the world to test, and you already built the fixtures in testdata/.

Every one of the four bugs above would have been caught by a small table test. I found them in minutes precisely because the components are well separated. Write these yourself — AGENTS.md is right that this is where the learning is. Start with the lexer: give it a source string, assert the token kinds and spans.

internal/report is dead code. PrintTerminalReport has zero callers. I grepped the whole tree. Meanwhile main.go:94-102 prints its own ad-hoc report inline, and printSymbol lives in main.go:105. You have two reporting implementations, and the better-looking one (with the ASCII logo) never runs. Pick one. Move printSymbol into report and call it, or delete the package.

Pos carries only a byte offset (tokens.go:3). You clearly intended more — a bare struct wrapping one int exists only to be extended. Without line and column, your HTML cannot link a symbol to its source line, which is the main thing readers want from generated docs. The lexer is the only place that can compute this cheaply, because it is the only place that sees every newline.

The CLI has no flags. The output path is hardcoded to output/out.html relative to the current working directory (main.go:79-83), not to the project you are documenting. Run it from two directories and you get two different output locations. Look at the flag package for -o and a quiet switch.

main.go does too much. It walks, lexes, parses, groups by directory, sorts, assembles, creates directories, renders, and prints. That is eight responsibilities in one function. Everything from line 47 to line 78 is pipeline orchestration. Could that become one function in a package, so main only parses arguments and reports errors?

Inconsistent handling of a fatal file. A lexer error calls os.Exit(1) and kills the whole run (main.go:56-57), but a load error just skips the file (main.go:51). Both are per-file problems. Why do they get different treatment?

defer f.Close() never runs on the error paths. os.Exit skips deferred calls (main.go:88 vs main.go:92). You also never check the error from Close, and for a written file that is where a disk-full error surfaces.

Small things
Git hygiene: .DS_Store and bin/godocgen are both tracked. Your .gitignore lists /bin/, but the file was committed before the rule existed, so the rule has no effect. git rm --cached is what you want. A committed 3 MB binary will bloat history.
consumeDigit has an operator precedence trap (lexer.go:90): a && b || cparses as(a && b) || c. The isValid()guard does not protect thepeek() == '.'test. It happens not to loop forever, becausepeekreturns0` past the end — but the code reads as if the guard covers both. Add parentheses to say what you mean.
.note is emitted but never styled. html.go:74 writes class="note", and the CSS block has no .note rule. Zero matches.
anchor can collide (html.go:80): it maps / to -, so a/b and a-b produce the same ID.
Two rules disagree on the package name. Assemble keeps the first MODULE (assemble.go:20), and AssemblePackage takes the last (assemble.go:72). On twin_a the tree says twina and the heading says twin_a. Only your invalid fixture triggers it, but the disagreement is real.
gofmt -l . fails on the intentionally-broken fixture. go vet and go build correctly ignore testdata/, but gofmt walks it. Worth knowing before you add CI.
Where I would start
Fix the doc-comment loss first — it corrupts output silently, and it is a one-line condition. Then make Assemble deterministic, since it is equally small. Then write lexer tests, and let those tests drive the bounds fix and the UTF-8 change, because both need cases you can point at.

The foundation here is genuinely good. The bugs are all in leaf functions, not in the structure — which is the far easier position to be in.

```
Great notes. So here’s the thing, a lot of the stdout is just debugging right now. Even the output/out.html is technically debugging. The goal is serve this with net/http stdlib. It’s all going to go soon. You gave me lots to work with though. I want to expand the languages to markdown and Python soon so this is useful for my repos. Then once those are good and I’m ready for pain I’m going to take on JS/Ts 

"Curtis"
```

Good — that reframes things. Let me recalibrate, and first verify one thing that changes severity a lot once this is served over HTTP.

Tested whether anchor output is escaped
Confirmed, and this one matters much more now:

```html
<a href="#pkg-evil"onmouseover=alert(1) x">
<section id="pkg-evil"onmouseover=alert(1) x">
```
anchor() never escapes (html.go:80). You escape the link text correctly but inject the attribute raw (html.go:49-51, html.go:66). Today the path comes from argv, so it is a curiosity. Serve this over HTTP against repos you did not write, and it is stored XSS.

Recalibrating the last review
Several of my points were aimed at the wrong target. Discard these:

CLI flags, hardcoded output/out.html, defer f.Close() — all dead on arrival. The file sink goes away.
internal/report and printSymbol — you already know it is scaffolding. Delete when the server lands.
.note CSS — the stylesheet is going to be rewritten anyway.
These get more important, not less:

The out-of-bounds read was a curiosity when you ran it on your own code. Over HTTP it is an attacker-controlled panic. net/http recovers the goroutine, but the client gets a truncated response and your log fills with stack traces.
Determinism is no longer about clean diffs. It is about ETag and Last-Modified. You cannot cache a page that reorders itself every render.
UTF-8 goes from edge case to daily breakage. Markdown is prose. Prose is full of non-ASCII.
One thing you got right that pays off now: Render(w io.Writer, ...) already accepts an http.ResponseWriter unchanged. The seam was in the right place.

The HTTP move
Answer this before you write a handler: do you parse per request, or build an index once?

Everything follows from it. Per-request parsing is simple and always fresh, but you re-lex the whole repo on every refresh. An index parsed at startup is fast and cacheable, but now you own invalidation and you have shared state across goroutines.

For a docs server over repos you control, I would build the index once and add an explicit refresh path. But make it a deliberate choice.

If you index once, the follow-up question is ownership. Your Symbol structs contain slices (Children). If handlers read a shared []PackageDoc concurrently, that is safe only while nothing mutates it. Is anything in your current Assemble path mutating in place? Look at assemble.go:32 — topLevel[i].Children = append(...). What does that do to a shared index if it runs twice?

Your sticky-error pattern becomes a liability. I praised pageWriter.firstError last time. Over HTTP it inverts. You write straight to the ResponseWriter, so the first Write commits a 200 OK and flushes headers. When Render returns an error at html.go:62, the client already has half a page and a success status. You cannot fix the status code after the fact.

Buffer the render into memory, check the error, then choose the status and copy out. What do you gain besides correct status codes? Think about Content-Length and ETag.

Move to html/template. Not for tidiness — for the bug above. html/template is context-aware: it knows an attribute from a text node from a URL from a JS string, and escapes each correctly. html.EscapeString does not know where its output lands, which is exactly why anchor() slipped through. Templates would have caught it with no thought from you.

If a URL selects a package or file, treat the path as hostile. filepath.Join(root, userInput) with ../../etc/passwd is the oldest bug in web serving. Look at filepath.Clean semantics carefully, and at what http.ServeFile and os.Root (Go 1.24+) do about it. Better: never map URL to filesystem path at all — key into your index by an ID you generated.

Where the language abstraction breaks
This is the interesting part, and your instinct to sequence Markdown/Python before JS/TS is right. But I think you have the shape wrong.

Sort your targets by lexical model, not by difficulty:

Language	Model	Fits your lexer?
Go	C-family, braces, free whitespace	yes, it is built for this
JS/TS	C-family, braces, free whitespace	yes, with work
Python	C-family tokens, significant indentation	partly
Markdown	line-oriented, structural, no tokens	no
Three things are hardcoded in the lexer core that must become table data:

String delimiters. `, ", ' are literal cases in the Tokenize switch (lexer.go:41-46). Python needs """ and prefixes like r" and f".
Comment syntax. // and /* are baked into consumePunctOrOper (lexer.go:126, lexer.go:133). Python uses #. Markdown uses <!-- -->.
Whitespace. Tokenize discards it (lexer.go:34-36). Python cannot.
Items 1 and 2 are the enabling refactor. They are mechanical, and they move rules from code into your existing Language struct — exactly the design you already committed to. Do them before you add any language.

Item 3 is different in kind. Python needs the lexer to emit INDENT and DEDENT tokens, which means tracking a stack of indent widths and synthesising tokens that correspond to no bytes in the source. What does that do to your Span? What is the span of a DEDENT?

The deeper problem: Owner is a Go-ism.

Go is unusual. Methods are declared at top level and name their receiver. So you must resolve a method to its type by name afterwards — which is why Assemble exists, and why "orphaned method" is even a concept.

Python and JS/TS do not work this way. A method sits lexically inside its class body. You know the parent structurally, at parse time. An orphan is impossible.

So assemble.go is not a shared pipeline stage. It is the Go backend's fixup pass. What happens to Assemble when three of your four languages do not need it?

Markdown does not belong in the lexer at all. It has no identifiers, no keywords, no operators, no braces. Its structure is headings and fences, and it is line-oriented. Every abstraction in lexer.go is wrong for it. Do not force it.

So where is the real interface? Not at the token level. At the symbol level — the boundary you already built:

type Extractor interface {
    Extract(src []byte, path string) ([]Symbol, []Note)
}
registry: ".go" -> goExtractor
          ".py" -> pyExtractor
          ".md" -> mdExtractor
Go, Python and JS/TS extractors share your tokenizer core with different Language tables. Markdown shares nothing and is maybe eighty lines of line scanning. Symbol and PackageDoc stay the contract, and render never learns a new language exists.

Two questions that will shape this:

Does SymbolKind grow per language (HEADING, CLASS, MODULE, DECORATOR), or do you map everything onto a shared vocabulary? You already left commented-out kinds in symbols.go:22-44, so you have been thinking about it. Which choice keeps render from growing a switch per language?
"Package" means a directory in Go, a file in Python, and nothing in Markdown. Your main.go:66 groups by directory, which is a Go assumption. What is the grouping unit when a repo has all four?
The order I would work in
UTF-8 in the lexer, now. It is a core change, and every language you add makes it more expensive. Markdown makes it mandatory.
collectDoc and the map-ordering fix. Both are small, and determinism is a prerequisite for HTTP caching.
Move comment and string syntax out of the lexer into Language. The enabling refactor. No new language yet.
Markdown second, not last. Counterintuitive, but do it because it does not fit. It forces the Extractor seam while the codebase is still ~1400 lines. If you do Python first you will bend Language to fit, it will mostly work, and then Markdown will demand a rewrite of a larger system.
Python. Now you find out what INDENT/DEDENT does to your span model.
The HTTP server, once you have two dissimilar languages proving the interface holds.
JS/TS last. You were right that it is the painful one — template literals with ${} interpolation need the lexer to recurse into itself, JSX is a different grammar mid-expression, and TypeScript types are their own language.
One more thought on step 6: you may want the server earlier than this, purely because reloading a browser is a much faster feedback loop than reading stdout. If so, build the crudest possible handler now — index once, one route, no caching — and treat it as scaffolding just like the current dump.

Write the lexer tests before step 3. That refactor moves behaviour from code into data, and it is exactly the kind of change where a table test tells you instantly whether Go still parses the way it did an hour ago.

```
I’ve been told Go interfaces are discovered not planned. You mentioned my architecture is solid. How do interfaces fit into this. I remember from bootcamp that if it quacks like a duck, it may as well be a duck. How do interfaces fit into the mix?

"Curtis"
```


Where the duck analogy holds and where it breaks
Your bootcamp instinct is half right.

Right: a type never declares that it implements anything. No implements keyword. No base class. If a type has the methods, it satisfies the interface. Nobody has to sign up.

Wrong: in Python, duck typing is dynamic. You find out it was not a duck when the call panics in production. Go is structural and static. The compiler checks the quacking at build time. You get the flexibility of duck typing with the safety of a type system.

The consequence is bigger than it sounds. Because the implementer never names the interface, you can define an interface that types you do not own already satisfy. You can write an interface today that a library from 2015 fits, without touching it. That is the property the whole idiom is built on.

The rule that actually matters
Define the interface in the package that consumes it, not the package that implements it.

This is where people coming from Java or TypeScript go wrong. There, the interface is a contract you publish and implementers import. In Go it is a requirement the caller states about its input.

io.Reader lives in io. Does os.File import io to satisfy it? No. os.File just has a Read method with the right shape. io.Reader was written by someone describing what they needed, not by someone describing what File is.

So the question that produces good Go interfaces is never "what is this thing?" It is "what is the smallest thing I need from my input?"

That is also why Go interfaces are tiny. io.Reader has one method. The proverb: the bigger the interface, the weaker the abstraction. An interface with eight methods can only ever be satisfied by the one type you wrote it for.

You have already done this twice
Here is the part I think will land. You did not plan a single interface, and your architecture is solid because of two you inherited from the standard library.

One: render.Render(w io.Writer, ...). You wrote it against *os.File. But you typed the parameter as the smallest thing you needed — something you can write bytes to. The result is that the renderer already works with http.ResponseWriter and with *bytes.Buffer in a test, with zero changes. That is not luck; that is the idiom working. You asked "what do I need?" instead of "what is this?"

Two: fs.FS, and here you stopped one inch short. Look at walker.go:28-30:

WalkFiles(rootPath string)          <- takes a path
    fileSystem := os.DirFS(rootPath)  <- builds the abstraction inside
You are already walking an fs.FS. But you construct it inside the function, so the abstraction is sealed in and the caller gets no benefit. What would change if WalkFiles took an fs.FS instead of a string?

Tests could use fstest.MapFS — an in-memory filesystem, no disk, no testdata directory, no cleanup.
Later you could serve docs straight out of a zip, an embedded FS, or a git object store.
One catch, and it is instructive: you would also have to change os.ReadFile(absPath) at walker.go:46 to read through the filesystem value instead. The two changes only make sense together. That is the sort of thing you only see by looking at working code — which is the whole argument for discovery.

This is my concrete suggestion for a first interface exercise. It is small, it is real, and you do not design anything. You notice you already fit an interface the standard library wrote for you.

Now the tension in what I told you
I proposed an Extractor interface last message, before you have written Python or Markdown. Fair challenge: is that not exactly the "designing interfaces up front" the proverb warns against?

Partly, yes. Let me sharpen it.

The proverb is advice for when you have no evidence. With one implementation, an interface is a guess about the future, and the guess is usually wrong in a way you cannot see yet.

You are not quite in that position — you have one working extractor and specific knowledge that Markdown differs structurally. But the honest version of my advice is:

Write the Markdown extractor as a plain concrete function first. Do not make it satisfy anything. Then put it beside the Go path and look at what they actually share. The interface falls out of the comparison. If you write the interface first, you will unconsciously bend Markdown to fit it, and you will learn nothing.

Concretely, here is what you cannot know until you have written the second one:

Does an extractor need path, or only src? Markdown may need it for relative links. Go does not.
Does it return an error? Your Go path never fails — it skips what it cannot parse. If only one language needs an error, that is a signal about the whole design, not about that language.
Does it take []byte or io.Reader? Your lexer indexes randomly into source to build spans. Streaming is impossible. That constraint is invisible from the outside and only shows up in the implementation.
Three real questions, none answerable by thinking harder today. The rule is not "never write an interface." It is that an interface with one implementation is a hypothesis, and an interface with two is an observation.

Where not to add interfaces
The usual overcorrection after this clicks is to make everything an interface. Do not.

Lexer and Parser should stay concrete structs. There is one lexer. Wrapping it in an interface adds a layer of indirection and buys you nothing — you cannot swap the implementation, and you have no second one to swap in.

And notice you do not need interfaces for testing here either. CreateParser(tokens []lexer.Token, src []byte) takes pure data. You can construct exactly the input you want and assert on exactly the output. That is strictly better than a mock. In Go, needing a mock is often a hint that a function is reaching for the world instead of accepting it as a parameter — which is the same lesson as the WalkFiles example, from the other direction.

The short version
Ask "what do I need from this input?", never "what is this thing?"
Put the interface where it is used, not where it is implemented.
Keep it to one or two methods. Big interfaces are weak ones.
Accept interfaces, return concrete structs.
Write the second implementation before you write the interface that covers both.
No implementation should ever import an interface only to satisfy it.
The reason your architecture holds up is that you kept your stage boundaries as plain data — tokens in, symbols out, PackageDoc to the renderer. Data is the most flexible contract there is, and interfaces are what you reach for only when one side of that boundary needs to vary. Right now exactly one side does: the front end, once Markdown arrives.

Everything else should stay concrete until a second implementation proves otherwise.
