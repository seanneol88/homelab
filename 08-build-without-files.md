# Building an empty room — `no configuration file` / `not found` / `No such container`

**Symptoms (the cascade, in order):**
```
failed to solve: … "/requirements.txt": not found
no configuration file provided: not found
Error response from daemon: No such container: python-api
```

**True meaning:** three error *generations* from one cause:
1. **Build context lacks the files** the Dockerfile references (COPY target)
2. Compose finds no `docker-compose.yml` in the cwd at all (even emptier)
3. The *echo*: no container exists because no image was ever built — said by
   a later command (`docker logs`) about the wreckage of 1-2

**Root cause (ours):** files were missing — created as directories (→ 07),
deleted by a repeat `rm` (→ 07b), or never saved by the editor (VS Code silent
loss — cured by terminal heredocs).

**Fix — Rule 0, engraved:**
> **Before `docker compose up`, run `ls` and count the files.**
> Files first, build second. Always.

```bash
ls -la /opt/stacks/<stack>/      # expect the full file set, all -rw- rows
cd /opt/stacks/<stack> && docker compose up -d --build
```

**The emergency bypass when editors misbehave** (proved twice this journey):
write all files via terminal in one paste —
```bash
cat > file << 'EOF'
…content…
EOF
```
Quoted `'EOF'` = paste literally, no variable expansion (`${…}` survives).

**Prevention:** reading errors **top-down** turns cascades into one line:
the first failure is the cause, the rest are echoes. When in doubt,
`docker compose build 2>&1 | tail -60`.

**First seen:** 2026-10-12 · python-api quest, three consecutive loops.
