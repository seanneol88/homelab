# Escaping the skeleton — JS-built UI renders markup as literal text

**Symptom:** after a JS-driven rebuild, the page prints raw source as
visible text —
```html
<div class="note"><span class="num">#3</span><div>I am sleeper</div>…
```
— every tag spelled out, the counter pill fine, data correct, styling dead.

## True meaning

The anti-XSS armor was wrapped around the **entire row template**:

```js
body.innerHTML = notes.map(n => esc(
  `<div class="note">…${n.text}…</div>`
)).join('');
```

`esc()` converts `<` / `>` / `&` into entities — *including the tags it
was protecting*. The browser receives soup of entities, so it has nothing
left to interpret as structure: the bones get displayed, the page never
renders. Infrastructure was fine all along (the Add button had even
written to the DB seconds before — the data layer was innocent).

## Root cause

v2 refactor created the delete-chip rows by extending the v1 template and
accidentally escalated the escaping scope from the *one dynamic field* to
the whole frame. Law violated:

> **Escape the DATA, never the SKELETON.** `esc()` may hug only
> user-provided values (note text, names, messages) — its own markup
> tokens must stay literal or the page self-strangles.

## Fix

Per-field armor, exactly one dynamic point wrapped:

```js
body.innerHTML = notes.map(n =>
  `<div class="note"><span class="num">#${n.id}</span><div>${esc(String(n.text))}</div>…</div>`
).join('');
```

Then the embedded-UI rebuild rite: rewrite file → `docker compose up -d
--build` → **Ctrl+Shift+R** in the browser (the served page is a compiled
binary artifact + a cached-browser artifact — both ghosts must be busted).

## Prevention / durable lessons

1. When a rendered page shows *source as text*, suspect escaping scope
   **before** touching the backend; one `grep -n "esc(" index.html` finds
   the hug.
2. Template-with-JS apps: keep the audit reflex — every `${…}` that
   carries user bytes gets `esc()`; the surrounding HTML never does.
3. If nano-level substring repairs drift past one round, **restore by
   canon** (whole-file heredoc rewrite) — determinism over surgery.
4. Prove the *served* build carries the fix with a fingerprint probe, no
   browser involved:
   ```bash
   curl -s https://<host>/ | grep -c "esc(String(n.text))"   # expect 1
   ```

**First seen:** 2026-10-13 · step-16, go-app delete-button refactor
(add-note #3 "I am sleeper" proved the write path survived unchanged).
