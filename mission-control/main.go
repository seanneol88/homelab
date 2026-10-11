package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ---------- data ----------

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Result struct {
	Service
	Up        bool
	LatencyMs int64
	CheckedAt string
}

var (
	mu      sync.RWMutex
	results []Result
)

// ---------- health checking ----------

func check(s Service) Result {
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.URL)
	up := err == nil && resp != nil && resp.StatusCode < 500
	if resp != nil {
		resp.Body.Close()
	}
	return Result{
		Service:   s,
		Up:        up,
		LatencyMs: time.Since(start).Milliseconds(),
		CheckedAt: time.Now().Format("15:04:05"),
	}
}

func pollAll(svcs []Service) {
	out := make([]Result, len(svcs))
	var wg sync.WaitGroup
	for i, s := range svcs {
		wg.Add(1)
		go func(i int, s Service) { // check all services at the same time
			defer wg.Done()
			out[i] = check(s)
		}(i, s)
	}
	wg.Wait()
	mu.Lock()
	results = out
	mu.Unlock()
}

// ---------- web page ----------

var page = template.Must(template.New("dash").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="30">
<title>Mission Control</title>
<style>
 body{margin:0;font-family:system-ui,sans-serif;background:#0d1117;color:#e6edf3;
      display:flex;flex-direction:column;align-items:center;padding:2rem}
 h1{margin:0}.sub{opacity:.6;margin:.4rem 0 2rem;font-size:.9rem}
 .grid{display:flex;gap:1rem;flex-wrap:wrap;justify-content:center;max-width:900px}
 .card{background:#161b22;border:1px solid #30363d;border-radius:12px;
       padding:1.1rem 1.4rem;min-width:200px}
 .card b{font-size:1.05rem}.card p{margin:.35rem 0 0;font-size:.85rem;opacity:.65;word-break:break-all}
 .up{color:#3fb950}.down{color:#f85149}
</style>
</head>
<body>
 <h1>&#128640; Mission Control</h1>
 <p class="sub">seanneol.com homelab &middot; auto-refresh 30s</p>
 <div class="grid">
 {{range .}}
  <div class="card">
   {{if .Up}}<span class="up">&#9679; UP</span>{{else}}<span class="down">&#9679; DOWN</span>{{end}}
   <b>&nbsp;{{.Name}}</b>
   <p>{{.URL}}</p>
   <p>{{.LatencyMs}} ms &middot; checked {{.CheckedAt}}</p>
  </div>
 {{end}}
 </div>
</body>
</html>`))

func dashboard(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	data := results
	mu.RUnlock()
	page.Execute(w, data)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ---------- main ----------

func main() {
	servicesFile := os.Getenv("SERVICES_FILE")
	if servicesFile == "" {
		servicesFile = "/app/services.json"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	raw, err := os.ReadFile(servicesFile)
	if err != nil {
		log.Fatal("cannot read services file: ", err)
	}
	var svcs []Service
	if err := json.Unmarshal(raw, &svcs); err != nil {
		log.Fatal("bad services file: ", err)
	}

	pollAll(svcs) // first check immediately
	go func() {   // then re-check every 30s in the background
		for range time.Tick(30 * time.Second) {
			pollAll(svcs)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", dashboard)
	mux.HandleFunc("GET /healthz", healthz)

		log.Printf("mission control listening on :%s, watching %d services", port, len(svcs))
	log.Fatal(http.ListenAndServe(":"+port, mux))
}





