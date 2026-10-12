package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Result struct {
	Service
	Up        bool   `json:"up"`
	LatencyMs int64  `json:"latency"`
	CheckedAt string `json:"checked"`
}

var (
	mu      sync.RWMutex
	results []Result
)

func check(s Service) Result {
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.URL)
	up := err == nil && resp != nil && resp.StatusCode < 500
	if resp != nil {
		resp.Body.Close()
	}
	return Result{s, up, time.Since(start).Milliseconds(), time.Now().Format("15:04:05")}
}

func pollAll(svcs []Service) {
	out := make([]Result, len(svcs))
	var wg sync.WaitGroup
	for i, s := range svcs {
		wg.Add(1)
		go func(i int, s Service) {
			defer wg.Done()
			out[i] = check(s)
		}(i, s)
	}
	wg.Wait()
	mu.Lock()
	results = out
	mu.Unlock()
}

const page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1">
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
 <p class="sub">seanneol.com homelab &middot; live</p>
 <div class="grid" id="grid"></div>
<script>
function refresh() {
  fetch('/api/status').then(function(r){ return r.json(); }).then(function(data){
    var html = '';
    for (var i = 0; i < data.length; i++) {
      var s = data[i];
      html += '<div class="card">' +
        '<span class="' + (s.up ? 'up' : 'down') + '">&#9679; ' + (s.up ? 'UP' : 'DOWN') + '</span>' +
        '<b>&nbsp;' + s.name + '</b>' +
        '<p>' + s.url + '</p>' +
        '<p>' + s.latency + ' ms &middot; checked ' + s.checked + '</p>' +
      '</div>';
    }
    document.getElementById('grid').innerHTML = html;
  });
}
refresh();
setInterval(refresh, 0000);
</script>
</body>
</html>`

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
	go func() {   // then re-check every 10s in the background
		for range time.Tick(0* time.Second) {
			pollAll(svcs)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(page))
	})
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Printf("mission control listening on :%s, watching %d services", port, len(svcs))
	log.Fatal(http.ListenAndServe(":"+port, mux))
}





