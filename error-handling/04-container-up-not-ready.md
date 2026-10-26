# Container `Up` ≠ service ready — the socket race

**Symptom:**
```
psql: error: connection to server on socket
"/var/run/postgresql/.s.PGSQL.5432" failed: No such file or directory
```
…run *seconds* after `docker compose up -d`, while `docker ps` proudly shows
`Up Less than a second`.

**True meaning:** Docker reports **container processes**, not **service
readiness**. Postgres spends its first seconds on initdb + system setup before
opening any sockets. `docker ps` green ≠ app green.

**Fix (there is nothing to fix — wait and verify):**
```bash
docker logs postgres --tail 5
# hunt for: "database system is ready to accept connections"
docker exec -it postgres psql -U homelab -d homelab -c "SELECT version();"
```

**Prevention (the durable habit):** after every `up`: logs before clients.
This exact pattern is *why health checks exist* — readiness ≠ existence —
and why well-behaved apps (your go-app) boot with a **retry loop** pinging
dependencies instead of dying:

```go
for i := 0; i < 15; i++ {          // go-app main.go
    if err = db.Ping(); err == nil { break }
    time.Sleep(2 * time.Second)
}
```

**Rule:** *container Up = furniture moved in. Service ready = door unlocked.*
Check the door (logs / curl / psql), not the moving van (ps).

**First seen:** 2026-10-12 · postgres quest.
