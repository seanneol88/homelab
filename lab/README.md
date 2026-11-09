<div align="center">

# 🐟⛩️ seanneol.com — the homelab kingdom

**one 4GB server · one domain · many small sturdy organs**

`Go ✕1 binary` · `k3s (single-node)` · `Traefik` · `cloudflared tunnel` · `FluxCD GitOps` · `Cloudflare Turnstile` · `Uptime Kuma` · `restic`

![traffic](https://img.shields.io/badge/traffic-pod--carried-7ee081) ![edge](https://img.shields.io/badge/edge-cloudflare-orange) ![ops](https://img.shields.io/badge/ops-git--push--only-blueviolet) ![status](https://img.shields.io/badge/status-lilac--armored-c9f34a)

</div>

---

## 🏛 The two faces

| door | wears | answers|
|---|---|---|
| **seanneol.com** | `seanneol · reading` — the hobby page (crowned 2026-11-08) | books & Goodreads |
| **www.seanneol.com** | same face | — |
| **lab.seanneol.com** | the **lab panel** — converter · docs · pocket AI | workbench for the world |

Beneath: **Traefik ingress** (class-bound) → **cloudflared** sings into the tunnel → **k3s pods (replicas:2, gitOps-managed)** carry the Groq/OpenAI brain. No port open on the VM; not even the control plane panel is exposed to the wild.

## 🧬 Stack

| organ | posture |
|---|---|
| **lab-lab v9** (Go, single binary) | `healthz` bare; `/api/*` lanes w/ rate-limit + Turnstile vault-check |
| **k3s** | namespace `lab`, secrets from env-file, probes on `/healthz` |
| **FluxCD** | repo = sovereign truth; two-drum pattern (`flux-system` + `lab-apps`) |
| **cloudflared** | local-managed tunnel @ `/etc/cloudflared/config.yml`; apex+www+lab routed `:80` |
| **Cloudflare Access** | Bypass-guest lanes · Allow-locked admin lanes only |
| **Cloudflare Turnstile** | invisible latch; strangers meet `{"error":"turnstile: prove humanity"}` |
| **Uptime Kuma** | arms on `lab` healthz + apex reading-face |
| **restic** | nightly vault; verify drift chorus-song |

## 🧭 Public routes

- `GET /healthz` → `{model, provider, status}` (bare JSON; monitored)
- `POST /api/chat` `/api/summarize` `/api/extract` `/api/convert` `/api/convert-doc` `/api/imagine` — **Turnstile-latched** (humans pass; bots 403)
- literature sired tonight: `homelab-systems/` (21 chapters) + `error-handling/` (24 scars) + `DISASTER-RUNBOOK.md`

## 🥁 Ops doctrine

```bash
edit-in-repo → git push → flux hears in ≤60s → pods morph  # hands never touch kubectl for change
poke-if-impatient: flux reconcile source git flux-system && flux reconcile kustomization lab-apps
freeze-for-surgery: flux suspend kustomization lab-apps …hands… flux resume kustomization lab-apps
undo-any-rite: git revert HEAD && git push
```

`image:` pins are immutable on the drum-side (`lab-lab:vN`); compose-lane stays the warm museum. 🌒

---

<div align="center">
<em>built by SeaNeolEbneazer & one night of chaos, crystallized · 2026 · seanneol.com</em>
</div>
