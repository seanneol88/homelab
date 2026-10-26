# `requires go >= 1.25.0` — floating dep vs pinned toolchain

**Symptom (Docker build):**
```
go: toolchain upgrade needed to resolve github.com/jackc/pgx/v5/stdlib
go: github.com/jackc/pgx/v5@v5.11.0 requires go >= 1.25.0
(running go 1.23.12; GOTOOLCHAIN=local)
```

**True meaning:** Build layer. `go mod tidy` resolved the dependency to the
*newest* version (`@latest`-style), and that fresh code demands a newer Go
than your builder image pins. New library refuses old compiler.

**Root cause:** the Dockerfile pinned `golang:1.23-alpine`; the pgx driver
marched forward between playbook-writing day and build day. Version churn —
normal Tuesday in software.

**Fix (bump both ends of the contract):**
```dockerfile
FROM golang:1.25-alpine AS build      # was 1.23
```
```gomod
go 1.25                               # go.mod, was 1.23
```
```bash
docker compose up -d --build
```

**Prevention:**
- Read build errors **top-down**: the first `go:` line was the diagnosis; the
  famous "No such container" at the end was just the echo (→ 08).
- Error messages that name exact versions are fix instructions wearing a scary
  costume: "requires go >= 1.25" = "set 1.25".
- Pins buy reproducibility; freshness buys security. When they fight, upgrade
  deliberately (this), don't shrug-pin.

**First seen:** 2026-10-12 · go-app build (pgx v5.11.0 vs Go 1.23).
