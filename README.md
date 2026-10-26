# 🏰 Seanneol fortress library

> One workspace, four archives, one codex — everything the homelab taught us, in verse and fact.
> _"Verify the artifact, never the memory."_

---

## 📦 The shelves

| folder | what it holds | format |
|---|---|---|
| 🗺 `homelab-systems/` | **the castle atlas** — every system's birth, operate-commands, secrets ledger, if-it-breaks maps | md chapters |
| 🧪 `lab-seanneol-com/` | the public organ-body mirror — lab's complete canon (Go python pod, docs engines, Dockerfile, 5 laws on the gate) | code + md |
| 📈 `skill-tree/` | workflow XP — STEPS 00–16 of becoming fortress-architect (habits, not systems) | md steps |
| 🧷 `error-handling/` | 19 scar-family templates — pattern injuries (paste-glue, template-faith, model-gerontology …) | md cases |
| 📜 `homelab-playbook/` | longform guides + `DISASTER-RUNBOOK.md` (BOOK that rebuilds the fortress from ash) | md |
| 🖼 `guides/` | visual guides (restic rotation HTML, etc.) | html |
| 🎒 bundles | `homelab-systems.tar.gz` · `skill-tree.tar.gz` · `error-handling.tar.gz` — tar snapshots for travel | gz |

---

## ⛩️ The gate laws (memorized — they save hours)

1. **Paste-glue** — multi-line `.env` from anywhere can fuse; cat five lines before any `up`.
2. **Template-faith** — `PASTE_…`/`KEEP_CURRENT_…` placeholders become REAL VALUES if swallowed.
3. **Model-gerontology** — hosted AI catalogs churn; `/v1/models` is the only oracle, every host.
4. **Alpine-µsl-shuns-C++-wheels** — pip journeys with opencv-class natives → glibc images (`python:3.12-slim-bookworm`).
5. **Builder-TTY-swallows-compiler** — "failed to solve" with no line shown → `--progress=plain` or compile outside docker.

---

## 🤝 Covenant

- New system born → say **"new chapter for `<system>`"** → the atlas grows, bundle re-tar'd.
- New scar earned → say **"new scar"** or just paste the error → error-handling gains the case.
- New habit proven → it marches to XP in skill-tree.
- Bundles refresh after every quilt-day (README + children fresh).

---

## 🌌 Tonight's trail (chronicle-tail)

- `lab.seanneol.com` — 3+ engines live (converter · groq-120b chat · flux imagine · docs: txt/docx/pptx + ✨ summarize)
- next dawn-quests: apex `seanneol.com` host-swap carries the hobby-reading page · Kuma casualty re-pointing · rotate leaked tokens · Step-17 scribe → when silence returns 💙🐟🔥

# lab.seanneol.com — the public homelab

> One Go binary · one tunnel · three+ engines: chat brain, pixel-conjurer, doc organs.

```txt
guest (internet)
  │
  ▼
Cloudflare edge   [DNS · TLS · Turnstile gate for seanneol.com]
  │
  ▼
cloudflared tunnel   hostname lab.seanneol.com
  │
  ▼
VM docker-server   LAN :8097  ->  container :8080
  │
  └─ lab (Go 1.23, one binary) serves:
       /api/convert       image png <-> jpg         (pure Go)
       /api/extract       pdf|docx|pptx|txt -> text (pdftotext / zip+XML)
       /api/convert-doc   pdf -> .docx / .pptx / .txt
       /api/chat          groq · openai/gpt-oss-120b (24-msg memory)
       /api/summarize     groq chunked map+reduce
       /api/imagine       cloudflare workers ai · flux-1-schnell
       /healthz           liveness (200/301/302 ok)

brains called per request:   Groq API · Cloudflare Workers AI
python organs in-container:  pdftotext · pdf2docx · pdftoppm + python-pptx
```

## Engines

| engine | endpoint | muscle |
|---|---|---|
| Convert (images) | `POST /api/convert` | pure-Go stdlib |
| Docs → text | `POST /api/extract` | pdftotext / stdlib-zip-OOXML |
| Docs → docx / pptx / txt | `POST /api/convert-doc` | pdf2docx · pdf2slides.py |
| Chat | `POST /api/chat` | **groq · openai/gpt-oss-120b** |
| Summarize | `POST /api/summarize` | groq chunked |
| Imagine | `POST /api/imagine` | CF flux-1-schnell |

## Config (`.env` untracked)

| var | effect |
|---|---|
| `CHAT_PROVIDER` | `groq` (default; others: cloudflare, arena) |
| `CHAT_MODEL` | `openai/gpt-oss-120b` — probe `/v1/models` when it rusts |
| `GROQ_API_KEY` | chat + summarize |
| `CF_ACCOUNT_ID` | CF Workers AI + cloudflare-chat |
| `CF_API_TOKEN` | Workers AI token (60-char family) |
| `CHAT_BUDGET` | default 1024 |
| `PORT` | 8080 inside container |

## Rituals

- Code bake: `docker compose up -d --build` (wheels+opencv cook ≈ 15 min the first fire)
- Env-only: `docker compose up -d --force-recreate`
- Truth probe: `docker exec lab env | sort`
- Build-mystery bypass: `--progress=plain` or `docker run --rm -v "$PWD:/app" -w /app golang:1.23-alpine sh -c "go build ."`

## The five laws (carved on the gate)

1. **Paste-glue** — multi-line `.env` can fuse; cat five lines before any `up`.
2. **Template-faith** — `PASTE_…` skeletons become REAL VALUES if swallowed.
3. **Model-gerontology** — catalogs churn; `/v1/models` is the only oracle.
4. **Alpine-µsl-shuns-C++-wheels** — pip+u slc natives → glibc runtime (`python:3.12-slim-bookworm`).
5. **Builder-TTY-swallows-compiler** — blind "failed to solve" → `--progress=plain`/compile outside.

# 🌳 Homelab Skill Tree — As-Built Record

> *seanneol.com empire · built 2026 · operator: seanneol88*
> The record of what was actually done — quests in order, skills earned, scars kept.

## XP Ledger

| Step | Quest | Core skills unlocked |
|---|---|---|
| 01 | Remote Workshop | 🐚 terminal · ssh · heredocs · VS Code Remote |
| 02 | Docker Foundations | 🐳 compose · volumes · ports · logs · exec |
| 03 | Git & GitHub | 🌿 add/commit/push · remotes · secrets hygiene |
| 04 | Tunnel + DNS | 🟧 cloudflared · ingress rules · systemd |
| 05 | Website | 🌐 static serving · host routing · forensics |
| 06 | Zero Trust Access | 🔐 identity wall · public-vs-admin design |
| 07 | Mission Control | 📡 health checks · polling loops |
| 08 | PostgreSQL + DBeaver | 🐘 DB server · SQL · secrets |
| 09 | Go Full-Stack API | 🏗️ Go HTTP · driver · JSON API |
| 10 | Live Dashboard | ⚡ in-place updates · fetch/JSON |
| 11 | Python API | 🐍 FastAPI · layer forensics · incident handling |
| 12 | Uptime Kuma | ⏰ monitors · docker DNS · Telegram alerts |
| 13 | Backups | 💾 restic · pg_dump · systemd timer · push-heartbeat |
| 14 | Disaster Runbook | 📜 trust topology · restore choreography |
| 15 | Public Status Page | 🌍 path-level trust zoning · Bypass policy |
| 16 | Full-Stack Faces | 🖼️ go:embed vs HTMLResponse · CRUD UI · /proc vitals · esc()-law |

**Rank: Full-Stack Homelaber III — Immortal** — design, deploy, expose, guard, store, watch, resurrect on hardware you own.

Every step ends with a ⛩️ Gate — the exact command/sight that proves the step. Never declare victory without passing it.
