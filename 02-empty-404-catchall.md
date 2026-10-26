# Empty `HTTP ERROR 404` — the cloudflared catch-all answering

**Symptom:** tunnel status green/connected, yet the browser shows a naked
`HTTP ERROR 404` with no content.

**Fingerprint:** **0-byte body**. An app-issued 404 (FastAPI etc.) carries JSON
like `{"detail":"Not Found"}` (~22 bytes). Empty = catch-all.

**True meaning:** Layer 3 (ingress). cloudflared fell through every
hostname rule and hit the final `- service: http_status:404`. The name you're
asking for isn't in the *running* config.

**Root causes (seen all three):**
1. The hostname rule was never added to `/etc/cloudflared/config.yml`
2. Rules placed **below** the catch-all (evaluation is top-down; catch-all eats everything beneath it)
3. Rules added but `systemctl restart cloudflared` skipped — process still holds the old config

**Fix:**
```bash
grep -n "<hostname>" /etc/cloudflared/config.yml    # is the rule written?
# if missing/misplaced → full heredoc rewrite (rules above; catch-all LAST)
systemctl restart cloudflared                       # always after config edits
curl -s https://<hostname>/healthz ; echo           # verify
```

**Prevention:** one canonical full-file heredoc for config.yml (never partial
appends); catch-all stays last; restart + verify = one motion.

**Sibling (the echo you never met):** if the *rule* exists but the local app is
down, the answer would be **502** — "tunnel works, origin doesn't". Then fix
layer 4-5 (`docker ps`, `curl localhost:<port>`), not layer 3.

**First seen:** 2026-10 · website 404 epic · re-fingerprinted during python-api.
