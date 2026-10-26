# INC-001 — the dashboard brain-freeze (first real production incident)

**Status:** 🟡 MITIGATED (restart), root cause UNCONFIRMED, recurrence watch active.

## Timeline (2026-10-12, UTC)

| Time | Event |
|---|---|
| 05:12 | mission-control rebuilt ("1s live ticker" version) and started |
| 05:13:30–31 | **last completed check round** — data freezes here |
| ~05:56 | first red flag: dashboard showing "DOWN Python API" in browser |
| ~06:05–06:20 | forensics: three channels (public /healthz 200 ✅, app /docs ✅, but brain shows 05:13 data) |
| ~06:26 | `docker restart mission-control` |
| 06:30:5x | fresh live data; **all five cards green** |

## What made it hard: the stale-photo red herring

First hypothesis was browser caching (the screenshots showed identical
timestamps — usually "old photograph"). **Elimination tool:** the control
group — VS Code's SSH port-forward gave direct, tunnel-free, Access-free
access: `localhost:8082/healthz` (app ✅ alive) and `localhost:8090`
(dashboard raw) — which rendered the *same frozen 05:13:30* straight from the
server's memory. Photo theory dead; the brain was truly stale.

## Mechanism (working theory, unproven)

The data pipeline: `pollAll` every 1s (ticker goroutine) → results →
`/api/status`. Serving stayed alive; **results stopped updating** ⇒ a pollAll
*round* never completed again — likely a check wedged during the
Cloudflare/config/DNS churn exactly at ~05:13 (Go App showed 993 ms just
before the freeze — jitter was spiking). Since each HTTP client carries a 5 s
timeout, a hard-hang shouldn't be possible… which is why the incident stays
**open-classified**.

## The response pattern (keep this — it's the SRE way)

1. **Unfreeze** (restart) — recovery > understanding, in the moment
2. **Capture evidence** (timestamps, sequence, `history`) — before it dissolves
3. **Watch for recurrence** — instrument *before* the next restart, not after

## If it happens again — the capture drill (do BEFORE restarting)

```bash
date -u                                          # true "now"
curl -s http://localhost:8090/api/status | head -c 400 ; echo   # stale stamps
docker logs mission-control --tail 20            # any whisper
docker exec mission-control ps aux               # process view
ss -tnp | grep mission                           # stuck connections?
```

Then instrument: add a log line per round inside `pollAll` ("round n: 5/5 ok,
took Xms") + one per check failure (service name + error), rebuild, and let it
happen again *observed*. The next freeze will name the wedged check itself.

## Lessons banked

- Monitors can fail in "serving but stale" mode — **trust timestamps, not presence**
- Control groups (localhost forwards) separate "app truth" from "pipe truth" in one move
- Stale-data bugs are monitors quietly monitoring their own failure — the most fun kind
- A documented open incident beats a perfectly-explained forgotten one

**First seen:** 2026-10-12 · python-api finale. File maintained under the same
name; update upon recurrence.
