# Timer never arms — `enable` ≠ started; and the UTC-human mismatch

**Symptom:** `systemctl list-timers homelab-backup.timer` prints only
"Pass --all to see loaded but inactive timers, too." — i.e. **no active
timer exists**. Overnight job never ran; manual runs work fine.

## True meaning (two co-conspirators)

1. **enable ≠ start.** `systemctl enable X.timer` only plants a symlink so
   the unit fires *at next boot*. A freshly-created timer needs
   `systemctl enable --now X.timer` (or an explicit `start`) — otherwise it
   sleeps until you reboot, which on a homelab may be months. Rite:
   create unit files → `daemon-reload` → `enable --now` →
   `list-timers` MUST be eyeballed for a NEXT column.

2. **The clock's citizenship.** Host was `Etc/UTC`; `OnCalendar=03:15`
   silently meant 03:15 *UTC* (11:15 local). Any scheduled math done in
   your head must be done in the server's TZ, or the server must learn
   yours: `timedatectl set-timezone Asia/Kuala_Lumpur`. Cron/timers/logs
   all then speak one language — yours.

## Prevention

- After ANY `systemctl (enable|start)`: one verification command
  (`list-timers` for timers, `is-active` for services). Five-second
  gate culture.
- Pin the host TZ once per new machine (first-boot rite) so every
  logline from here on is in your own time.
- Expect the heartbeat canary to lag: with a 25 h window, a missed day
  alarms ~a day late — cross-check with `restic snapshots` when hunting.

**First seen:** 2026-10-13 · backups quest (morning-after verification).
