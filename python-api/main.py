from datetime import datetime
from zoneinfo import ZoneInfo

from fastapi import FastAPI

app = FastAPI(title="homelab-api", version="0.1.0")


@app.get("/")
def root():
    return {"service": "api.seanneol.com", "status": "running"}


@app.get("/healthz")
def healthz():
    return {"status": "ok"}


@app.get("/api/time")
def current_time():
    now = datetime.now(ZoneInfo("Asia/Kuala_Lumpur"))
    return {"timezone": "Asia/Kuala_Lumpur", "iso": now.isoformat()}
