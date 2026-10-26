# `[42P01] relation "notes" does not exist`

**Symptom (DBeaver):** `SQL Error [42P01]: ERROR: relation "notes" does not
exist · Position: 15` — raised by `SELECT * FROM notes;` in a fresh script.

**True meaning:** SQL-layer truthfulness: the table genuinely doesn't exist —
because **only the SELECT ran**. The `CREATE TABLE` above it was never executed.

**Root cause:** DBeaver's *two* execute buttons:
- plain ▶ / **Ctrl+Enter** = run ONLY the statement under the cursor
- solid orange ▶ / **Alt+X** = run the whole script (or the selection, if any)

Cursor was parked on the last line → single-statement mode executed the SELECT
alone.

**Fix:** run in dependency order — click `CREATE TABLE` → Ctrl+Enter →
`INSERT` → Ctrl+Enter → `SELECT` → Ctrl+Enter. (Or select-all + Alt+X.)

**Related sibling:** `ERROR: relation "notes" already exists` — the opposite
end of the same knife: CREATE re-run on an existing table. Harmless: skip to
INSERT/SELECT. Both errors boil down to **statement ordering**.

**Prevention:** SQL is sequential state, not a document you "run." Know which
button runs how much; when a script has build-then-query steps, run it
top-to-bottom.

**First seen:** 2026-10-12 · postgres quest, first-ever DBeaver script.
