package main

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

//go:embed static/hobby.html
var hobbyHTML []byte
//go:embed static/reading-nook.jpg
var readingNookJPG []byte
const chatSystem = "You are a helpful assistant living on seanneol lab (lab.seanneol.com), a homelab showcase. Be concise, friendly, and honest. Never claim to have abilities you do not have."
const summarizeSystem = "You are a precise document summarizer. Output: 1) a 2-3 sentence TL;DR, 2) bullet key points, 3) notable entities/figures if any. Be faithful to the source; do not invent facts. If the excerpt is cut mid-phrase, ignore the fragment."

type chatReq struct {
	Message   string      `json:"message"`
	Model     string      `json:"model"`
	SessionID string      `json:"sessionId"`
	Budget    int         `json:"budget"`
	History   []chatEntry `json:"history"`
}
type chatEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatResp struct {
	Reply    string `json:"reply"`
	Provider string `json:"provider"`
}

type summarizeReq struct {
	Text string `json:"text"`
}
type summarizeResp struct {
	Summary  string `json:"summary"`
	Chunks   int    `json:"chunks"`
	Provider string `json:"provider"`
	Chars    int    `json:"chars"`
}

// ---------------- doc text extraction ----------------

func extractText(name string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".md", ".csv", ".log", ".json":
		return string(data), nil
	case ".docx":
		return ooxmlText(data, func(s string) bool { return s == "word/document.xml" })
	case ".pptx":
		return ooxmlText(data, func(s string) bool {
			return strings.HasPrefix(s, "ppt/slides/slide") && strings.HasSuffix(s, ".xml")
		})
	case ".pdf":
		return pdfText(data)
	default:
		return "", fmt.Errorf("unsupported document type %q (accept: pdf, docx, pptx, txt, md, csv, log, json)", ext)
	}
}

// DOCX/PPTX: unzip, filter members, strip XML tags, join the character data.
func ooxmlText(data []byte, keep func(string) bool) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid OOXML zip: %v", err)
	}
	var b strings.Builder
	for _, f := range zr.File {
		if !keep(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		raw, _ := io.ReadAll(io.LimitReader(rc, 8<<20))
		rc.Close()
		dec := xml.NewDecoder(bytes.NewReader(raw))
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			switch t := tok.(type) {
			case xml.CharData:
				b.Write(t)
			case xml.StartElement:
				if t.Name.Local == "br" || t.Name.Local == "p" {
					b.WriteByte('\n')
				}
			}
		}
		b.WriteString("\n\n")
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("no text found inside document")
	}
	return out, nil
}

func pdfText(data []byte) (string, error) {
	tmp, err := os.CreateTemp("", "lab-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "pdftotext", "-layout", tmp.Name(), "-").CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("pdf extraction timed out")
	}
	s := strings.TrimSpace(string(out))
	if err != nil && s == "" {
		return "", fmt.Errorf("pdftotext failed: %v", err)
	}
	if s == "" {
		return "", fmt.Errorf("pdf contained no extractable text (scanned image?)")
	}
	return s, nil
}

// ---------------- provider plumbing ----------------

func cfURL(account, model string) string {
	return fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", account, model)
}

type cfChat struct {
	Messages []chatEntry `json:"messages"`
	MaxToks  int         `json:"max_tokens,omitempty"`
}
type cfResp struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
	Raw     json.RawMessage
	Body    string
}

func callCloudflare(model string, e chatEntry, tok, acc string) (string, error) {
	body, _ := json.Marshal(cfChat{Messages: []chatEntry{{Role: "system", Content: chatSystem}, e}})
	req, _ := http.NewRequest("POST", cfURL(acc, model), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudflare unreachable: %v", err)
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if r.StatusCode != 200 {
		return "", fmt.Errorf("cloudflare said %d: %.300s", r.StatusCode, raw)
	}
	var c cfResp
	if json.Unmarshal(raw, &c) != nil || c.Result.Response == "" {
		return "", fmt.Errorf("cloudflare answered oddly: %.300s", raw)
	}
	return c.Result.Response, nil
}

const groqURL = "https://api.groq.com/openai/v1/chat/completions"

func callGroqMessages(model, apiKey string, sys string, entries []chatEntry, budget int) (string, error) {
	msgs := []chatEntry{{Role: "system", Content: sys}}
	msgs = append(msgs, entries...)
	body, _ := json.Marshal(map[string]any{
		"model": model, "messages": msgs, "max_tokens": budget,
	})
	req, _ := http.NewRequest("POST", groqURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	r, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("groq unreachable: %v", err)
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if r.StatusCode != 200 {
		return "", fmt.Errorf("groq brain said %d: %.300s", r.StatusCode, raw)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(raw, &out) != nil || len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("groq answered oddly: %.300s", raw)
	}
	return out.Choices[0].Message.Content, nil
}

func callArena(prompt, sys, key string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"model": "claude-sonnet-4.6", "system": sys, "prompt": prompt, "pending": "false",
	})
	req, _ := http.NewRequest("POST", "https://api.ai.cloudflare.com/v1/chat", bytes.NewReader(body))
	if key != "" {
		req.Header.Set("Authorization", "ApiKey "+key)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("arena unreachable: %v", err)
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if r.StatusCode != 200 {
		return "", fmt.Errorf("arena said %d: %.300s", r.StatusCode, raw)
	}
	var a struct {
		Reply string `json:"reply"`
	}
	if json.Unmarshal(raw, &a) != nil || a.Reply == "" {
		return "", fmt.Errorf("arena answered oddly: %.300s", raw)
	}
	return a.Reply, nil
}

func callArenaAnthropic(prompt string, maxToks int, key string) (string, error) {
	body := fmt.Sprintf(`{"model":"claude-sonnet-4.6","max_tokens":%d,"system":%s,"messages":[{"role":"user","content":%s}]}`,
		maxToks, jsonString(chatSystem), jsonString(prompt))
	req, _ := http.NewRequest("POST", "https://api.ai.cloudflare.com/v1/messages", bytes.NewReader([]byte(body)))
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	r, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("arena unreachable: %v", err)
	}
	defer r.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if r.StatusCode != 200 {
		return "", fmt.Errorf("arena said %d: %.300s", r.StatusCode, raw)
	}
	var a struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(raw, &a) != nil || len(a.Content) == 0 {
		return "", fmt.Errorf("arena answered oddly: %.300s", raw)
	}
	return a.Content[0].Text, nil
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func verifyTurnstile(tok, ip string) bool {
    secret := os.Getenv("TURNSTILE_SECRET")
    if secret == "" { return true }
    if tok == "" { return false }
    resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify",
        url.Values{"secret": {secret}, "response": {tok}, "remoteip": {ip}})
    if err != nil { return false }
    defer resp.Body.Close()
    var out struct{ Success bool `json:"success"` }
    if json.NewDecoder(resp.Body).Decode(&out) != nil { return false }
    return out.Success
}

func withTurnstile(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !verifyTurnstile(r.Header.Get("X-Turnstile-Token"), r.RemoteAddr) {
            w.Header().Set("Content-Type", "application/json")
            http.Error(w, `{"error":"turnstile: prove humanity"}`, http.StatusForbidden)
            return
        }
        next(w, r)
    }
}

// ---------------- HTTP handlers ----------------

func withLimit(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
		h(w, r)
	}
}

func chatHandler(prov, model string, budget int) http.HandlerFunc {
	type message struct {
		ts    time.Time
		entry chatEntry
	}
	var history = map[string][]message{}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var in chatReq
		if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Message) == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "send {\"message\":\"...\"}"})
			return
		}
		sid := in.SessionID
		if sid == "" {
			sid = "anon"
		}
		now := time.Now()
		entries := history[sid]
		var recent []chatEntry
		for _, m := range entries {
			if now.Sub(m.ts) < 2*time.Hour {
				recent = append(recent, m.entry)
			}
		}
		if len(recent) > 24 {
			recent = recent[len(recent)-24:]
		}
		for _, hmsg := range in.History {
			if hmsg.Role == "user" || hmsg.Role == "assistant" {
				if len(hmsg.Content) > 4000 {
					hmsg.Content = hmsg.Content[:4000]
				}
				recent = append(recent, hmsg)
			}
		}
		entry := chatEntry{Role: "user", Content: in.Message}
		btv := budget
		if in.Budget > 0 && in.Budget < 2000 {
			btv = in.Budget
		}

		var reply string
		var err error
		var used = prov
		switch prov {
		case "groq":
			key := os.Getenv("GROQ_API_KEY")
			if key == "" {
				reply = "my groq brain is unpowered — GROQ_API_KEY is empty in the lab's env."
				err = nil
				break
			}
			userEntries := append([]chatEntry{}, recent...)
			userEntries = append(userEntries, entry)
			reply, err = callGroqMessages(model, key, chatSystem, userEntries, btv)
			used = "groq"
		case "arena":
			reply, err = callArena(in.Message, chatSystem, os.Getenv("ARENA_API_KEY"))
			used = "arena /v1/chat"
			if err != nil && os.Getenv("ARENA_API_KEY") != "" {
				if alt, err2 := callArenaAnthropic(in.Message, btv, os.Getenv("ARENA_API_KEY")); err2 == nil {
					reply, err, used = alt, nil, "arena /v1/messages"
				}
			}
		default:
			tok := os.Getenv("CF_API_TOKEN")
			acc := os.Getenv("CF_ACCOUNT_ID")
			if tok == "" || acc == "" {
				reply = "my cloudflare brain is unplugged — CF_API_TOKEN/CF_ACCOUNT_ID are empty."
				err = nil
			} else {
				reply, err = callCloudflare(model, entry, tok, acc)
				used = "cloudflare"
			}
		}
		if err != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "provider": used})
			return
		}
		history[sid] = append(entries, message{now, entry}, message{now, chatEntry{Role: "assistant", Content: reply}})
		json.NewEncoder(w).Encode(chatResp{Reply: reply, Provider: used})
	}
}

func summarizeHandler(prov, model string, budget int) http.HandlerFunc {
	const chunkSize = 12000
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var in summarizeReq
		if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Text) == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "send {\"text\":\"...document text...\"}"})
			return
		}
		if prov != "groq" || os.Getenv("GROQ_API_KEY") == "" {
			w.WriteHeader(503)
			json.NewEncoder(w).Encode(map[string]string{"error": "summarizer needs CHAT_PROVIDER=groq with a key"})
			return
		}
		key := os.Getenv("GROQ_API_KEY")
		btv := budget
		chunks := 1
		var summary string
		var err error
		runes := []rune(in.Text)
		if len(runes) <= chunkSize {
			summary, err = callGroqMessages(model, key, summarizeSystem,
				[]chatEntry{{Role: "user", Content: string(runes)}}, btv)
		} else {
			var partials []string
			for i := 0; i < len(runes); i += chunkSize {
				end := i + chunkSize
				if end > len(runes) {
					end = len(runes)
				}
				p, e2 := callGroqMessages(model, key,
					"Summarize the KEY facts of this document part in tight bullets.",
					[]chatEntry{{Role: "user", Content: string(runes[i:end])}}, 512)
				if e2 != nil {
					err = e2
					break
				}
				chunks++
				partials = append(partials, p)
				if len(partials) >= 25 {
					break
				}
			}
			if err == nil {
				summary, err = callGroqMessages(model, key, summarizeSystem,
					[]chatEntry{{Role: "user", Content: "Combine these part-summaries into the final document summary:\n\n" + strings.Join(partials, "\n---\n")}}, btv)
			}
		}
		if err != nil {
			w.WriteHeader(502)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "provider": "groq"})
			return
		}
		json.NewEncoder(w).Encode(summarizeResp{Summary: summary, Chunks: chunks, Provider: "groq", Chars: len(runes)})
	}
}

func extractHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, hdr, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "multipart form field \"file\" required"})
			return
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, 32<<20))
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "could not read your upload"})
			return
		}
		text, err := extractText(hdr.Filename, data)
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "name": hdr.Filename})
			return
		}
		if len([]rune(text)) > 400000 {
			text = string([]rune(text)[:400000])
		}
		json.NewEncoder(w).Encode(map[string]any{
			"name": hdr.Filename, "chars": len([]rune(text)), "text": text,
		})
	}
}

func convertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "multipart form field \"file\" required", 400)
			return
		}
		defer f.Close()
		src, _, err := image.Decode(f)
		if err != nil {
			http.Error(w, "could not decode image (png/jpg/gif only)", 400)
			return
		}
		format := r.FormValue("to")
		width := atoi(r.FormValue("width"), 0)
		dst := src
		if width > 0 && src.Bounds().Dx() > width {
			dst = resize(src, width)
		}
		switch format {
		case "jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Disposition", `attachment; filename="converted.jpg"`)
			jpeg.Encode(w, dst, &jpeg.Options{Quality: 90})
		case "png":
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Disposition", `attachment; filename="converted.png"`)
			png.Encode(w, dst)
		default:
			http.Error(w, `"to" must be jpg or png`, 400)
		}
		_ = hdr
	}
}

func imagineHandler() http.HandlerFunc {
	type req struct {
		Prompt string `json:"prompt"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in req
		if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Prompt) == "" {
			http.Error(w, `send {"prompt":"...","width":1024,"height":1024}`, 400)
			return
		}
		tok := os.Getenv("CF_API_TOKEN")
		acc := os.Getenv("CF_ACCOUNT_ID")
		if tok == "" || acc == "" {
			http.Error(w, "imagine engine unplugged — CF_API_TOKEN/CF_ACCOUNT_ID empty", 503)
			return
		}
		payload, _ := json.Marshal(map[string]any{
			"prompt": in.Prompt, "steps": 4,
		})
		url := cfURL(acc, "@cf/black-forest-labs/flux-1-schnell")
		hr, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
		hr.Header.Set("Authorization", "Bearer "+tok)
		hr.Header.Set("Content-Type", "application/json")
		rsp, err := (&http.Client{Timeout: 120 * time.Second}).Do(hr)
		if err != nil {
			http.Error(w, "cloudflare unreachable", 502)
			return
		}
		defer rsp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(rsp.Body, 64<<20))
		if rsp.StatusCode != 200 {
			http.Error(w, fmt.Sprintf("cloudflare said %d: %.400s", rsp.StatusCode, body), 502)
			return
		}
		var j map[string]json.RawMessage
		if err := json.Unmarshal(body, &j); err == nil {
			w.Header().Set("Content-Type", "image/png")
			if s, ok := j["result"]; ok {
				var ri struct {
					Image string `json:"image"`
				}
				json.Unmarshal(s, &ri)
				if ri.Image != "" {
					b, _ := toBytes(ri.Image)
					w.Write(b)
					return
				}
			}
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="imagine.png"`)
		w.Write(body)
	}
}

func toBytes(b64 string) ([]byte, error) {
	const enc = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var out []byte
	acc, bits := 0, 0
	for _, c := range b64 {
		i := strings.IndexRune(enc, c)
		if i < 0 {
			continue
		}
		acc = acc<<6 | i
		bits += 6
		if bits >= 8 {
			bits -= 8
			out = append(out, byte(acc>>bits))
		}
	}
	return out, nil
}

func atoi(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

func resize(src image.Image, width int) image.Image {
	b := src.Bounds()
	h := width * b.Dy() / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, width, h))
	for y := 0; y < h; y++ {
		for x := 0; x < width; x++ {
			sx := x * b.Dx() / width
			sy := y * b.Dy() / h
			dst.Set(x, y, src.At(b.Min.X+sx, b.Min.Y+sy))
		}
	}
	return dst
}

func main() {
	prov := os.Getenv("CHAT_PROVIDER")
	if prov == "" {
		prov = "cloudflare"
	}
	model := os.Getenv("CHAT_MODEL")
	if model == "" {
		model = "@cf/meta/llama-3.1-8b-instruct"
	}
	budget := atoi(os.Getenv("CHAT_BUDGET"), 1024)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		h := strings.ToLower(r.Host)
		if strings.HasPrefix(h, "seanneol.com") || strings.HasPrefix(h, "www.seanneol.com") {
			w.Header().Set("Content-Type", "text/html")
			w.Write(hobbyHTML)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "provider": prov, "model": model})
	})
	http.HandleFunc("/api/chat", withLimit(withTurnstile(chatHandler(prov, model, budget))))
	http.HandleFunc("/api/summarize", withLimit(withTurnstile(summarizeHandler(prov, model, budget))))
	http.HandleFunc("/api/extract", withLimit(withTurnstile(extractHandler())))
	http.HandleFunc("/api/convert", withLimit(withTurnstile(convertHandler())))
	http.HandleFunc("/api/convert-doc", withLimit(withTurnstile(convertDocHandler())))
	http.HandleFunc("/api/imagine", withLimit(withTurnstile(imagineHandler())))
	http.HandleFunc("/reading-nook.jpg", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(readingNookJPG)
})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("lab listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
// ---------------- document conversion ----------------

func convertDocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "multipart form field \"file\" required", 400)
			return
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, 32<<20))
		if err != nil {
			http.Error(w, "upload too large or unreadable", 400)
			return
		}
		to := r.FormValue("to")
		base := strings.TrimSuffix(filepath.Base(hdr.Filename), filepath.Ext(hdr.Filename))

		if to == "txt" {
			text, err := extractText(hdr.Filename, data)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="`+base+`.txt"`)
			w.Write([]byte(text))
			return
		}

		if strings.ToLower(filepath.Ext(hdr.Filename)) != ".pdf" {
			http.Error(w, "docx/slides conversion requires a .pdf input", 400)
			return
		}
		tmp, err := os.CreateTemp("", "lab-doc-*.pdf")
		if err != nil {
			http.Error(w, "temp file failed", 500)
			return
		}
		src := tmp.Name()
		defer os.Remove(src)
		if _, err := tmp.Write(data); err != nil {
			tmp.Close()
			http.Error(w, "temp write failed", 500)
			return
		}
		tmp.Close()

		var out string
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		var cmd *exec.Cmd
		switch to {
		case "docx":
			out = src + ".docx"
			cmd = exec.CommandContext(ctx, "python3", "-m", "pdf2docx", "convert", src, out)
		case "pptx":
			out = src + ".pptx"
			cmd = exec.CommandContext(ctx, "python3", "/app/tools/pdf2slides.py", src, out)
		default:
			http.Error(w, `form value "to" must be txt, docx, or pptx`, 400)
			return
		}
		defer os.Remove(out)

		if blob, err := cmd.CombinedOutput(); err != nil {
			http.Error(w, fmt.Sprintf("converter said: %v :: %.300s", err, blob), 422)
			return
		}
		conv, err := os.ReadFile(out)
		if err != nil {
			http.Error(w, "conversion produced no file (heavy layout?)", 422)
			return
		}
		mime := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		if to == "pptx" {
			mime = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", `attachment; filename="`+base+"."+to+`"`)
		w.Write(conv)
	}
}
