from datetime import datetime
from zoneinfo import ZoneInfo

from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI(title="homelab-api", version="0.1.1")


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
.card{background:#0d1117;border:1px solid #21262d;border-radius:14px;padding:22px;margin-top:14px}
.tz{color:#7d8590;font-size:12px;letter-spacing:.18em;text-transform:uppercase}
.clock{font-size:64px;font-weight:700;font-variant-numeric:tabular-nums;letter-spacing:1px;margin-top:4px}
.date{color:#7d8590;font-size:14px}
.sync{color:#3fb950;font-size:12px;margin-top:10px}
.door{display:flex;gap:14px;align-items:baseline;background:#0d1117;border:1px solid #21262d;border-radius:12px;padding:12px 14px;margin-top:10px}
.door code{color:#58a6ff;font-size:13px;min-width:150px}
.door span{color:#7d8590;font-size:13px}
.door a{color:#58a6ff;font-size:13px;text-decoration:none}
a.status{font-size:13px;color:#58a6ff;text-decoration:none}
</style></head>
<body><main>
  <h1>🐍 seanneol api <span class="pill" id="doors">3 doors</span><span style="flex:1"></span>
    <a class="status" href="https://uptime.seanneol.com" target="_blank">status ↗</a></h1>
  <div class="sub">FastAPI in a container on my own Proxmox box · pure Python, same soil</div>
  <div class="card">
    <div class="tz">Asia/Kuala_Lumpur</div>
    <div class="clock" id="clock">··:··:··</div>
    <div class="date" id="date"></div>
    <div class="sync" id="sync">syncing against /api/time…</div>
  </div>
  <div class="door"><code>GET /</code><span>this face</span></div>
  <div class="door"><code>GET /api/time</code><span>KL time as JSON — the clock above is reading it</span></div>
  <div class="door"><code>GET /api/status</code><span>service heartbeat as JSON</span><a href="/api/status" target="_blank" style="margin-left:auto">open ↗</a></div>
  <div class="door"><code>GET /docs</code><span>interactive explorer (FastAPI's free gift)</span><a href="/docs" target="_blank" style="margin-left:auto">open ↗</a></div>
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
sync(); setInterval(sync, 8000);
paint(); setInterval(paint, 200);
</script>
</body></html>
"""


@app.get("/", response_class=HTMLResponse)
def root():
    return INDEX_HTML
