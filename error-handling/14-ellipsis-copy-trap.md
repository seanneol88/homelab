# The `…` copy-trap — visible truncation copied as truth

**Symptom:** API calls fail with 404/403 to an endpoint that LOOKS right —
`https://host/api/push/z2wc…?status=up…` — because the URL/stdin literally
contains the character `…` (U+2026), typed from the UI's display.

## True meaning

Many UIs shorten long secrets/tokens *on screen* with an ellipsis for layout.
That `…` is a rendering hint to human eyes: *"more characters live here."*
Selecting + copying the visible text (triple-click / drag) copies the
**presentation**, not the **value**. The machine downstream has no notion of
"there were more characters" — it just sees a wrong token → 404/403,
often with no readable clue (that's the cruel part).

## Fix / prevention

1. For any long-value field in a web UI: use the dedicated **copy button 📋**
   (copies the underlying value), never selection-based copying of a
   truncated display.
2. When in doubt, `grep`/echo the pasted value back and look with your eyes
   for `…`/`*` before running: *"does this string contain a character I
   wouldn't expect in a token?"*
3. Family tree of copy-hazards (same root law: *presentation ≠ value*):
   - ellipsis truncation (this file)
   - smart quotes/curly quotes from rich editors
   - trailing whitespace/newlines riding a paste
   - leading `Bearer `/`Authorization:` headers copied into a *token* field
   - timestamps hitchhiking off chat/message bubbles (see 12-case lore: the
     `05:50 PM` token suffix)
4. Law of the five-second check: **paste → eyeball → THEN execute.**

**First seen:** 2026-10-12 · backups quest (Kuma Push URL wiring; 404).
