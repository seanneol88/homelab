import os
from datetime import datetime
from zoneinfo import ZoneInfo

from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI(title="homelab-api", version="0.2.1")


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/api/time")
def current_time():
    now = datetime.now(ZoneInfo("Asia/Kuala_Lumpur"))
    return {"timezone": "Asia/Kuala_Lumpur", "iso": now.isoformat()}


@app.get("/api/status")
def api_status():
    return {"service": "api.seanneol.com", "status": "running"}


def _uptime():
    s = float(open("/proc/uptime").read().split()[0])
    d, rem = divmod(int(s), 86400)
    h, rem = divmod(rem, 3600)
    m, _ = divmod(rem, 60)
    return s, f"{d}d {h}h {m}m"


def _load():
    p = open("/proc/loadavg").read().split()
    return {"1m": float(p[0]), "5m": float(p[1]), "15m": float(p[2])}


def _mem():
    info = {}
    for line in open("/proc/meminfo"):
        k, v = line.split(":")
        info[k] = int(v.strip().split()[0])  # kB
    total = info["MemTotal"]
    avail = info.get("MemAvailable", info["MemFree"])
    return {"total_mb": total // 1024, "used_pct": round(100 * (1 - avail / total), 1)}


def _disk():
    st = os.statvfs("/")
    total = st.f_blocks * st.f_frsize
    free = st.f_bavail * st.f_frsize
    return {"used_pct": round(100 * (1 - free / total), 1), "free_gb": round(free / 1e9, 1)}


@app.get("/api/vitals")
def vitals():
    s, pretty = _uptime()
    return {
        "uptime_seconds": int(s),
        "uptime_pretty": pretty,
        "load": _load(),
        "memory": _mem(),
        "disk": _disk(),
        "sampled_at": datetime.now(ZoneInfo("Asia/Kuala_Lumpur")).isoformat(),
    }


INDEX_HTML = """<!DOCTYPE html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>seanneol · API</title>
<style>
:root{color-scheme:dark}
*{box-sizing:border-box}
body{margin:0;font:15px/1.55 -apple-system,"Segoe UI",Roboto,Arial,sans-serif;background:#0b0f14;color:#e6edf3}
main{max-width:860px;margin:0 auto;padding:48px 20px 80px}
h1{font-size:22px;letter-spacing:.4px;display:flex;align-items:center;gap:10px}
.pill{background:#12331f;color:#3fb950;border-radius:999px;padding:2px 10px;font-size:13px;font-weight:600}
.sub{color:#7d8590;font-size:13px;margin:6px 0 26px}
.section{color:#7d8590;font-size:12px;letter-spacing:.18em;text-transform:uppercase;margin:34px 4px 4px}
.card{background:#0d1117;border:1px solid #21262d;border-radius:14px;padding:22px;margin-top:14px}
.tz{color:#7d8590;font-size:12px;letter-spacing:.18em;text-transform:uppercase}
.big{font-size:34px;font-weight:700;margin:4px 0 2px;font-variant-numeric:tabular-nums}
.clock{font-size:64px;font-weight:700;font-variant-numeric:tabular-nums;letter-spacing:1px;margin-top:4px}
.date{color:#7d8590;font-size:14px}
.sync{color:#3fb950;font-size:12px;margin-top:10px}
.lbl{color:#7d8590;font-size:12px;margin-top:14px}
.bar{height:8px;background:#0b0f14;border:1px solid #21262d;border-radius:999px;margin-top:6px;overflow:hidden}
.bar>i{display:block;height:100%;background:#238636;transition:width .4s,background .4s}
.chips{display:flex;gap:8px;flex-wrap:wrap;margin-top:14px}
.chip{background:#0b0f14;border:1px solid #21262d;border-radius:999px;padding:3px 12px;font-size:12px;color:#7d8590}
.door{display:flex;gap:16px;align-items:center;background:#0d1117;border:1px solid #21262d;border-radius:12px;padding:14px 16px;margin-top:10px;transition:border-color .12s}
.door:hover{border-color:#30363d}
.door code{color:#d2a8ff;font-size:13px;min-width:120px;flex:none}
.door .what{color:#e6edf3;font-size:13px;flex:1}
.door a{color:#58a6ff;font-size:13px;text-decoration:none;flex:none}
.door a:hover{text-decoration:underline}
a.status{font-size:13px;color:#58a6ff;text-decoration:none}
</style></head>
<body><main>
  <h1>🐍 seanneol api <span class="pill">4 doors</span><span style="flex:1"></span>
    <a class="status" href="https://uptime.seanneol.com" target="_blank">status ↗</a></h1>
  <div class="sub">FastAPI in a container on my own Proxmox box · pure Python, same soil</div>
  <div class="card">
    <div class="tz">Asia/Kuala_Lumpur</div>
    <div class="clock" id="clock">··:··:··</div>
    <div class="date" id="date"></div>
    <div class="sync" id="sync">syncing against /api/time…</div>
  </div>
  <div class="card">
    <div class="tz">house vitals · docker-server</div>
    <div class="big" id="up">booted … ago</div>
    <div class="lbl" id="mem-lbl">memory …</div>
    <div class="bar" id="mem"><i></i></div>
    <div class="lbl" id="disk-lbl">disk …</div>
    <div class="bar" id="disk"><i></i></div>
    <div class="chips" id="loadchips"></div>
  </div>
  <div class="section">doors</div>
  <div class="door"><code>GET /docs</code><span class="what">try every endpoint in the browser — no curl needed</span><a href="/docs" target="_blank">open ↗</a></div>
  <div class="door"><code>GET /api/vitals</code><span class="what">the pulse JSON that feeds the vitals card</span><a href="/api/vitals" target="_blank">open ↗</a></div>
  <div class="door"><code>GET /api/status</code><span class="what">one-line service heartbeat</span><a href="/api/status" target="_blank">open ↗</a></div>
  <div class="door"><code>GET /api/time</code><span class="what">KL time as JSON — the clock's source</span><a href="/api/time" target="_blank">open ↗</a></div>
  <div class="door"><code>GET /</code><span class="what">this page</span><a href="/" target="_blank">open ↗</a></div>
</main>
<script>
const TF = new Intl.DateTimeFormat('en-MY',{timeZone:'Asia/Kuala_Lumpur',hour:'2-digit',minute:'2-digit',second:'2-digit',hour12:true});
const DF = new Intl.DateTimeFormat('en-MY',{timeZone:'Asia/Kuala_Lumpur',weekday:'long',day:'numeric',month:'long',year:'numeric'});
let offset = 0;
async function sync(){
  const el = document.getElementById('sync');
  try{
    const r = await fetch('/api/time', {cache:'no-store'});
    const j = await r.json();
    offset = Date.parse(j.iso) - Date.now();
    el.textContent = 'synced to /api/time · drift ' + offset + 'ms · resyncing every 8s';
  }catch(e){ el.textContent = 'sync failed: ' + e.message; }
}
function paint(){
  const t = new Date(Date.now() + offset);
  document.getElementById('clock').textContent = TF.format(t);
  document.getElementById('date').textContent = DF.format(t);
}
function bar(id, pct, label){
  const b = document.getElementById(id);
  b.querySelector('i').style.width = pct + '%';
  b.querySelector('i').style.background = pct > 85 ? '#f85149' : pct > 70 ? '#d29922' : '#238636';
  document.getElementById(id + '-lbl').textContent = label;
}
async function vitals(){
  try{
    const r = await fetch('/api/vitals', {cache:'no-store'});
    const v = await r.json();
    document.getElementById('up').textContent = 'booted ' + v.uptime_pretty + ' ago';
    bar('mem', v.memory.used_pct, 'memory ' + v.memory.used_pct + '% of ' + (v.memory.total_mb/1024).toFixed(1) + ' GB');
    bar('disk', v.disk.used_pct, 'disk ' + v.disk.used_pct + '% used · ' + v.disk.free_gb + ' GB free');
    document.getElementById('loadchips').innerHTML = ['1m','5m','15m'].map(k =>
      `<span class="chip">load ${k}&nbsp; ${v.load[k]}</span>`).join('');
  }catch(e){}
}
sync(); setInterval(sync, 8000);
paint(); setInterval(paint, 200);
vitals(); setInterval(vitals, 8000);
</script>
</body></html>
"""


@app.get("/", response_class=HTMLResponse)
def root():
    return INDEX_HTML
