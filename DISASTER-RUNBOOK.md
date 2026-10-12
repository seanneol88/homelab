# 🆘 DISASTER RUNBOOK — box dies → empire reborn

> **Situation:** docker-server (192.168.0.228) is dead beyond saving.
> Everything below rebuilds it in ~15–30 minutes of calm, ordered work.
> Print-polish: follow sequentially, tick each gate before moving on.
> Last verified against reality: 2026-10-13.

---

## 0. What saves you (no purchase possible on the bad day)

- [ ] This repo: `github.com/seanneol88/homelab` (clone target)
- [ ] Restic repository in R2 bucket `homelab-backup` (nightly, encrypted)
- [ ] Vault (Google Password Manager) entries:
  - `r2.cloudflarestorage.com` — S3 Access Key ID (user) + Secret (pass) + endpoint (note)
  - `restic-backups.local` — RESTIC_PASSWORD
  - `github.com`, `cloudflare.com` accts, Proxmox root, `uptime-kuma admin`, telegram bot (to restore alerts fast)
- [ ] Patience: never skip a ✅ gate to "save time".

---

## 1. Provision a new stage

- [ ] Fresh Debian 12 machine/LXC (2cpu/4GB/32GB), static LAN IP (old one
      fine if it's free), hostname `docker-server`
- [ ] `apt update && apt upgrade -y && apt install -y curl ca-certificates git`
- [ ] Docker: `curl -fsSL https://get.docker.com | sh`
- [ ] `timedatectl set-timezone Asia/Kuala_Lumpur`
- [ ] SSH access for yourself (VS Code Remote later)
✅ Gate: `docker run hello-world` prints.

## 2. Bring the memory home

- [ ] `mkdir /opt && cd /opt && git clone https://github.com/seanneol88/homelab.git stacks`
      (authenticate via the **github.com** vault entry / token if private)
- [ ] Recreate the **`.env` files that git deliberately forgets** — values live in
      the vault: `postgres/.env` (DB_PASSWORD), any stack `.env`s your notes mention
✅ Gate: `cd /opt/stacks && git log --oneline -3` shows recent history; `.env` files exist
      with `chmod 600`.

## 3. Reconnect the edge (tunnel)

- [ ] Cloudflare dash → Zero Trust → Networks → Tunnels
- [ ] Option A (token preserved in dashboard): copy `home-server` tunnel's
      install command → run `cloudflared service install <TOKEN>` on the new host
- [ ] Option B: create fresh tunnel → set Public Hostnames
      (website/app/api/home/status/portainer/proxmox…) to the same local targets
✅ Gate: tunnel shows **Healthy**; `https://status.seanneol.com` answers the PIN wall
      (after step 5 it will answer the app).

## 4. Pull the archived world back

```bash
apt install -y restic
mkdir -p /opt/backups
cat > /opt/backups/restic.env << 'EOF'
   # FILL 4 values from vault entries (S3 pair + repo URL + repo password)
EOF
chmod 600 /opt/backups/restic.env && source /opt/backups/restic.env
restic snapshots            # ✅ repo reachable, beads listed
restic restore latest --target /
```
- [ ] Populate volumes exactly as restored: kuma data under
      `/var/lib/docker/volumes/uptime-kuma_kuma-data` (restore handles
      paths verbatim; docker volumes are recreated "pre-populated" if you
      restore BEFORE first `up`)
✅ Gate: `ls /opt/backups/dumps` shows the latest `homelab-YYYY-MM-DD.sql`.

## 5. Wake the stacks (chainsaw order)

```bash
cd /opt/stacks/postgres      && docker compose up -d   # 1st: the vault keeper
docker exec postgres pg_dump -U homelab homelab >/dev/null   # reachability ping
cd /opt/stacks/mission-control && docker compose up -d
cd /opt/stacks/uptime-kuma   && docker compose up -d
cd /opt/stacks/website && up; go-app && up; python-api && up   # per-stack loops
cd /opt/stacks/portainer     && docker compose up -d
```
✅ Gate: `docker ps` — all `up`; `https://status.seanneol.com` serves Kuma UI
      (old monitors + history intact, because kuma_data was restored).

## 6. Restore the database content

```bash
docker exec -i postgres psql -U homelab homelab < /opt/backups/dumps/homelab-LATEST.sql
```
✅ Gate: `docker exec -it postgres psql -U homelab homelab -c 'select count(*) from notes;'`
      matches expectation; Go/Python cards in Mission Control green via /data.

## 7. Re-arm the night watch

- [ ] Recreate the systemd pair (they're part of /opt/backups restore… but
      units live in /etc! → copy from the runbook repo copy below, or re-create
      with the canonical heredocs from skill-tree step-13)
- [ ] `systemctl daemon-reload && systemctl enable --now homelab-backup.timer`
- [ ] `systemctl list-timers | grep homelab-backup` shows NEXT
- [ ] Volley once: `systemctl start homelab-backup.service` → kuma push monitor 🟢
✅ Gate: timer armed + a manual run lands a bead + Kuma push monitor green.

## 8. Certificates-of-life review

- [ ] All five hostnames curl 200 (public trio) + PIN-walled pair
- [ ] DBeaver connects on 192.168.0.228:5432
- [ ] Telegram: `docker stop python-api` → 🔴 arrives (~1 min) → `docker start` → 🟢
- [ ] `pct set 100 --onboot 1` re-checks (armor on the NEW stage's host too)
- [ ] Commit: `git add -A && git commit -m "resurrection 2026-××-×× verified" && git push`

---

### Philosophy, filed once

- Materials ≠ assembly. This page is the assembly.
- The vault is the root of trust; the repo + R2 are the body; this page joins them.
- Every gate exists because somebody once skipped it.
