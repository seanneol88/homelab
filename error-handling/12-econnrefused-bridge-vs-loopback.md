# `connect ECONNREFUSED 172.17.0.1:8090` — the bridge door, the loopback wall

**Symptom:** Uptime Kuma monitor (inside its container) aimed at
`http://172.17.0.1:8090/healthz` reports **DOWN** with
`connect ECONNREFUSED 172.17.0.1:8090` — while the app is provably alive.

## True meaning / fingerprint lesson

`ECONNREFUSED` = **the packet ARRIVED at a real address, and nobody listened
ON THAT EXACT address:port.** Instant answer, milliseconds. (Contrast:
timeout = packet vanished — firewall drop / black hole / nothing routed.)

Why nobody listened: Mission Control publishes `127.0.0.1:8090:8090` — bound
to the host's **loopback only**. Traffic arriving at the host via any *other*
host address (like `172.17.0.1`, docker0's gateway IP) intentionally finds no
docker-proxy. **127.0.0.1-binding is scoped, not host-wide.** Portainer's
`172.17.0.1:9443` succeeds only because *it* publishes `0.0.0.0:9443`.

## Root cause (design clash, not bug)

Loopback-only binding (our hardening on step 02) × sibling-container consumer
(Kuma) = unreachable by port-forwarding. Container-to-container traffic should
never go through host ports at all.

## Fix — Docker-native, zero loosening

Put Kuma on Mission Control's network; address it **by name** (embedded DNS
on user-defined bridges resolves container names):

compose addition (uptime-kuma):
```yaml
    networks: [ mc-net ]
networks:
  mc-net:
    external: true
    name: mission-control_default
```
monitor URL: `http://mission-control:8090/healthz`

Pre-verify before touching the UI:
```bash
docker exec uptime-kuma node -e "fetch('http://mission-control:8090/healthz').then(r=>r.text()).then(console.log)"
# => ok
```

## Prevention / durable lessons

1. **Match the addressing to the audience:**
   - host-local consumers (cloudflared on systemd) → `127.0.0.1` published ports ✅
   - container consumers → shared network + container **name** ✅
   - LAN clients → `0.0.0.0` (or explicit IP) + firewall rules ✅
2. `localhost` inside a container = *that container*, never the host. Packets
   to `172.17.0.1` reach the host's *bridge* address, not its loopback.
3. Know the three truths: refused vs timeout vs NXDOMAIN name three different
   failure actors.

**First seen:** 2026-10-12 · uptime-kuma quest (Mission Control monitor).
