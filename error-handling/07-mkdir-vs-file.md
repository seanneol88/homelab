# The `mkdir` family — files born as folders, folders eaten twice

## 7a — "mkdir made my requirements.txt a FOLDER"

**Symptom:** silent wrongness, discovered later —
```bash
mkdir -p /opt/stacks/python-api/requirements.txt
```
created a **directory named requirements.txt**. No error message. Every
downstream step then failed one level deeper.

**True meaning:** `mkdir` = make **directory**. Always. Files are made by
editors (VS Code New File) or by shells writing content (`cat > file`,
`tee file`). The tool doesn't know you "meant" a file.

**Detect (one-glance skill):** first character of `ls -l` output:
```
d rwxr-xr-x  → d = DIRECTORY
- rw-r--r--  → - = regular FILE
```

## 7b — the nuke-twice

**Symptom:** files existed (proved: `-rw- … requirements.txt`), then `rm -rf`
was re-run "just to be sure" — and ate them. Build then failed on emptiness.

**True meaning:** `rm` is idempotent in effect but not in blast radius: the
second run deletes *newer* content silently. Destructive commands have no
"are you sure" in a terminal.

**Common prevention for both:**
1. **Destructive commands run exactly once**, and only after `ls -la` shows
   exactly what will be eaten
2. *Files* ← `cat > name << 'EOF'` heredoc · *folders* ← `mkdir -p path`
3. Suspicion → `ls -la`. It is the cheapest truth serum in the toolbox.

**First seen:** 2026-10-12 · python-api quest (bash history lines 213–215 were
the flight-recorder evidence).
