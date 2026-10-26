# "Log in to <my-app>" on my own public door — Allow ≠ Bypass

**Symptom:** after creating path-scoped Access apps meant to be public, incognito
visitors meet a **"Cloudflare Access — Log in to status-public"** page; curl
witnesses 302 → a waterfall of base64 ending in
`…redirect_url=%2Fapi%2Fstatus-page%2Fheartbeat-list%2Fmain`.
Browser meanwhile shows shell-but-grey (`N/A` bars) when a session cookie
covers the HTML but not the APIs.

## True meaning

The policy attached to the new apps existed — but its **Action was `Allow`,
not `Bypass`** (dropdown default survived). Allow = "identify yourself, then
pass"; Bypass = "identity check skipped for this path." A logged-in browser
cookie can therefore make a *Allow-world look half-public* (shell passes,
APIs choke for strangers) — the perfect confusion engine.

## Fix

Access controls → Policies → open the reusable policy (`Public`) →
**Action: Bypass** → Save. One statute flipped; every app citing it healed at
once. Verify via the wall-tarot (deliberately SAFE to run unauthenticated):

```bash
curl -s -o /dev/null -w "%{http_code} %{redirect_url}\n" <url>
# 200/404/etc. = wall absent (404 means you reached the APP itself)
# 302 → *.cloudflareaccess.com* = still walled
```

## Prevention / durable lessons

1. After ANY Access change: incognito + curl 302-check. Cookie'd windows lie.
2. Read the evaluation law printed in the UI: "default-deny; Bypass and
   Service Auth evaluate first."
3. Slug-vs-name confusion (barrens `main` vs `default`) — checker URLs use
   the machine name (slug), never the display name.
4. Policy reuse is strength AND risk-multiplier: one statute, many citations
   — audit its action twice before saving.

**First seen:** 2026-10-13 · public status page quest (`/status/` public,
cloud grey until flip).
