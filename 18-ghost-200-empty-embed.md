# Ghost 200 — go:embed orphan var, and "no matching files found"

**Symptom (flavor A):** UI route answers `HTTP/2 200` — title bar loads,
page renders *blank*, and `curl -s <url> | wc -c` whispers **0 bytes**.
Everything logs happy; the container "Started". Nothing is "broken", yet
nothing arrives.

**Symptom (flavor B):** after the directive *is* present, the image
suddenly refuses to build:
```
main.go:20:12: pattern static/index.html: no matching files found
```

## True meaning

`//go:embed` is a **compiler incantation bound by adjacency** to the var
directly beneath it. Two separate sins, two separate poisons:

**A — the orphan var.** `var indexHTML []byte` sits alone (directive
missing or detached by a blank line/comment). Perfectly legal Go: it's
just an *empty slice*. The handler `w.Write(indexHTML)` therefore serves
a 200 with zero bytes — **the 200 lies because status ≠ payload.**

**B — the context-starved builder.** The directive now demands the raw
file *at compile time*, inside the build context — but the Dockerfile
only copied source:
`COPY go.mod main.go ./` sends nothing under `static/` to `/src`. The
compiler votes: "no matching files found". (`go:embed`'s promise — no
runtime file — tempts you to forget this; it still needs the file *while
baking*.)

## Fix

A: directive hugging var, no air between them:
```go
//go:embed static/index.html
var indexHTML []byte
```

B: the second plank in the build stage:
```dockerfile
COPY go.mod main.go ./
COPY static ./static/
```

Verification pair (browser-free, cache-proof):
```bash
curl -s https://<host>/ | head -6        # <!DOCTYPE html> must bite
curl -s https://<host>/ | wc -c          # > 0 — bytes are the truth
```

## Prevention / durable lessons

1. **The 200-with-0-bytes tarot**: a UI route that 200s void says
   "handler alive, body empty" — suspect the resource the handler writes
   (embed var, template, file read) before anything network-shaped.
2. Embed surgery = *three-piece graft*: import `_ "embed"` + directive-
   hugged var + Dockerfile COPY plank. Miss one, the ghost arrives.
3. `grep`-based anatomy check *before* the build can pre-empt both sins:
   ```bash
   grep -n "go:embed\|COPY" main.go dockerfile…   # presence + adjacency eyeball
   ```
4. Scar-law adjacency echo: comments/directives think in *paragraphs* —
   never let a blank line sit between an attribute and its subject.

**First seen:** 2026-10-13 · step-16, go-app full-stack face build
(flavor A detected by grep+wc; flavor B predicted and then confirmed by
the buildkit error line above the `failed to solve` echo).
