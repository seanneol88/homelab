# `DNS_PROBE_POSSIBLE` / "DNS address could not be found"

**Symptom (Chrome):** `This site can't be reached — <host>'s DNS address could
not be found` · code `DNS_PROBE_POSSIBLE`.

**True meaning:** Layer 1 — **the name doesn't resolve at all**. Nothing below
DNS exists yet in the lookup's eyes. (Distinct from the empty-404, which means
"name resolved but nothing matched".)

**Root causes:**
1. The DNS record was never created (`route dns` skipped or failed)
2. The record *was* created minutes ago, but **negative caching**: Windows +
   Chrome cached the earlier "doesn't exist" answer and keep serving the ghost

**Fix:**
```bash
# ground truth, bypassing every cache (on the server):
nslookup api.seanneol.com 1.1.1.1
# resolves to 104.21… / 172.67… = DNS EXISTS → cache problem
# NXDOMAIN = record truly missing → create it:
cloudflared tunnel route dns 29f51fc5-a210-4cf3-9c64-4245005008bb api.seanneol.com
```

Cache eviction on the **laptop** (PowerShell on Windows, not the server!):
```
ipconfig /flushdns
```
then a **new incognito window** (Chrome caches DNS per-process — same-tab
reload keeps the grudge).

**Dashboard tell:** a DOWN result in ~**30–40 ms** = instant-fail = DNS-class;
real service failures are slower (they at least try).

**Prevention:** after every new hostname, the ceremony is: `route dns` →
"Added CNAME" → nslookup ground truth → *then* browser. Caches expire alone
in minutes; don't fight them, route around them (incognito).

**First seen:** 2026-10-12 · python-api quest (api.seanneol.com).
