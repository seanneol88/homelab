# 🩹 Error Handling — The Scar Archive

> Every file here is an error that **really happened** on docker-server, with the
> exact symptom, what it actually meant, the fix that worked, and the habit that
> prevents its return. Sorted chaos into knowledge. New scars → new files,
> same template.

---

## The Five Laws (read before any debugging session)

1. **Read the actual error line, not the follow-on noise.** `docker logs … --tail 50`
   beats a screenshot of the last 3 lines. The *first* failure is the cause;
   "No such container" is just the echo of "build failed". **First failure wins.**
2. **Name the layer before touching anything.** DNS → tunnel → ingress →
   container → app. Find the lowest failing layer; never fix above it.
   (`DNS_PROBE_POSSIBLE` is layer 1; don't debug layer 5.)
3. **Fingerprint responses, don't vibes them.** Empty 404 = cloudflared
   catch-all. 22-byte JSON 404 = the app. 000 = nothing listening. Every
   error has a fingerprint; learn to take them.
4. **Idempotent fixes can be re-safely applied.** Heredoc full-file rewrites,
   `up -d --build`, `route dns` — designed to be re-run. **Destructive
   commands (`rm`) run exactly once, after `ls` confirms what's being eaten.**
5. **Green first, champagne after.** Gates (curls, dashboards, logs) are the
   contract; the victory commit is the signature. Never in reverse order.

## Quick-lookup index

| You see… | Open |
|---|---|
| `code: 1003` on `route dns` | `01-error-1003-record-exists.md` |
| Browser `HTTP ERROR 404`, empty page, tunnel "up" | `02-empty-404-catchall.md` |
| `DNS_PROBE_POSSIBLE` / "address could not be found" | `03-dns-probe-possible.md` |
| psql `No such file or directory` on the socket | `04-container-up-not-ready.md` |
| `[42P01] relation "x" does not exist` | `05-error-42p01-relation-missing.md` |
| `requires go >= 1.xx` during build | `06-toolchain-version-conflict.md` |
| Folder appeared where a file should be / files vanish after `rm` | `07-mkdir-vs-file.md` |
| `no configuration file provided` · `"/file": not found` · `No such container` | `08-build-without-files.md` |
| bash mangled something with `<`, `>`, or pipes by surprise | `09-placeholder-redirection.md` |
| "nothing to commit" · "up to date" · other scary-but-calm lines | `10-benign-messages.md` |
| Dashboard data frozen while app alive (timestamps stale) | `11-inc-001-dashboard-freeze.md` |
| `connect ECONNREFUSED <ip>:<port>` from a sibling container | `12-econnrefused-bridge-vs-loopback.md` |
| `bash: export: '...': not a valid identifier` on sourcing an .env | `13-env-file-mangled.md` |
| 404/403 to a URL whose pasted text contains `…` | `14-ellipsis-copy-trap.md` |
| Scheduled job silent; `list-timers` shows nothing | `15-timer-never-armed-enable-now.md` |
| "Log in to <app>" stalls a public path; curl 302s to cloudflareaccess | `16-allow-vs-bypass-access-policy.md` |
| Page renders literal `<div…>` tags as text after a JS rebuild | `17-escape-data-never-skeleton.md` |
| Root returns 200 with an empty body; or `pattern static/index.html: no matching files found` | `18-ghost-200-empty-embed.md` |

## The universal template (copy for new entries)

```markdown
# <Error name> — <fingerprint>
**Symptom:** the exact message, verbatim.
**True meaning:** what layer/actor is really complaining.
**Root cause:** what actually happened to us.
**Fix:** the exact commands that cured it.
**Prevention:** the rule written into muscle memory.
**First seen:** date · quest/context.
```
