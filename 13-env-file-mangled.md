# `source: not a valid identifier` — the half-dressed export (nano mishap)

**Symptom:**
```bash
$ source /opt/backups/restic.env
bash: export: `746ca81a728de3e60d06891f328b6079': not a valid identifier
$ restic init
Fatal: unable to open S3 backend: Key ID ($AWS_ACCESS_KEY_ID) is empty
```

## True meaning

One line of the env file had lost its `export NAME="` clothing (nano delete-
overshoot while replacing a placeholder) and was left as a bare hex string.
Shell reads it as "assign something to 746ca…" → illegal identifier → that
line aborts; typically downstream vars never load either. Every consumer
behaves as if no secrets were provided at all (confusing second symptom:
"Key ID is empty").

## Fix / prevention

1. Each line's shape is sacred: `export WORD_WORD="value"` — after any manual
   edit, eyeball the shapes before saving.
2. Re-source-test immediately:
   ```bash
   unset AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY RESTIC_PASSWORD RESTIC_REPOSITORY
   source /opt/backups/restic.env
   echo "key: ${AWS_ACCESS_KEY_ID:0:4}… | secret: ${AWS_SECRET_ACCESS_KEY:+YES} | pass: ${RESTIC_PASSWORD:+YES}"
   ```
   `${VAR:+YES}` proves presence without spilling.
3. `cat -A file` shows stray CR/quotes if Windows-style gremlins ever appear.
4. Env files hold secrets → `chmod 600`; never commit (already gitignored
   habits from step 09/11 apply).

**First seen:** 2026-10-12 · backups quest (restic.env arming).
