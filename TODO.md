# godocgen — Roadmap

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
