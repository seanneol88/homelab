# Error 1003 — "record already exists" on `cloudflared tunnel route dns`

**Symptom:**
```
An A, AAAA or CNAME record already exists with that host. (Code: 1003)
```

**True meaning:** DNS layer. Cloudflare refuses to overwrite an existing record —
you must consciously delete the old one. A protection, not a bug.

**Root cause (ours):** re-routing a hostname (e.g. `www`) that still had an old
CNAME — in the worst case, pointing at a *ghost tunnel* (`eb446493-…`), so the
site chained root → www → dead tunnel. Double misdirection.

**Fix:**
1. dash.cloudflare.com → seanneol.com → **DNS → Records** → find the row for
   that hostname → **Delete**
2. Re-run:
   ```bash
   cloudflared tunnel route dns 29f51fc5-a210-4cf3-9c64-4245005008bb <hostname>
   ```
   → expect `Added CNAME …`
3. No cloudflared restart needed — DNS records don't touch the tunnel process.

**Prevention:** before `route dns` on a hostname that has *ever* existed,
glance at the DNS table. One stale record hijacks a whole hostname.

**First seen:** 2026-10 · website quest (www) · then re-met on api. DNS table
ended with 6 healthy Tunnel-type records.
