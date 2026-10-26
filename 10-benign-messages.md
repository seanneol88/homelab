# Scary-calm messages — things that sound wrong but are right

A category all its own: output that *looks* like failure and means "already
handled." Memorize these; they return an hour of your life each time.

| Message | True meaning | Correct action |
|---|---|---|
| `nothing to commit, working tree clean` | your commit already landed (maybe earlier); nothing new to record | `git log --oneline -3` to admire; move on |
| `Your branch is up to date with 'origin/main'` | push already succeeded | none |
| `api.seanneol.com is already configured to route to your tunnel` | the DNS record exists from the first run | none — proof, not error |
| `mkdir: cannot create directory … File exists` (without `-p`) | folder already there | use `mkdir -p`, or just proceed |
| `Added CNAME …` (again) | re-adding is accepted/no-op'd | none |
| cloudflared 404 on unknown hostnames | catch-all doing its designed job for *other* (random bot) requests | it's the bouncer turning away strangers |

**The general law:** messages divide into **statements of fact** (calm,
present/perfect tense, no error code) and **cries for help** (error: / failed /
did not complete). Learn to tell them apart *before* acting — acting on calm
messages re-breaks working things (see 07b: "just to be sure"-driven `rm`).

**Debug reflex to build:** when unsure whether output is an error, look for
the literal word `error` / `failed` / exit codes (`echo $?` → `0` = success),
or verify with the next gate (curl / logs / ps) instead of the void.

**First seen:** everywhere, daily. The class that never graduates.
