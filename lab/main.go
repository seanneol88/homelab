package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

func jsonOut(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func httpPostJSON(url, auth string, extraHdr map[string]string, payload []byte, timeout int) ([]byte, int, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	for k, v := range extraHdr {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func cfRun(w http.ResponseWriter, model string, payload []byte) ([]byte, bool) {
	accID := os.Getenv("CF_ACCOUNT_ID")
	token := os.Getenv("CF_API_TOKEN")
	if accID == "" || token == "" {
		jsonOut(w, 500, map[string]string{"error": "cloudflare not wired — .env needs CF_ACCOUNT_ID + CF_API_TOKEN"})
		return nil, false
	}
	url := "https://api.cloudflare.com/client/v4/accounts/" + accID + "/ai/run/" + model
	body, code, err := httpPostJSON(url, "Bearer "+token, nil, payload, 120)
	if err != nil {
		jsonOut(w, 502, map[string]string{"error": "upstream unreachable: " + err.Error()})
		return nil, false
	}
	if code != 200 {
		jsonOut(w, 502, map[string]string{"error": fmt.Sprintf("cloudflare said %d: %s", code, string(body))})
		return nil, false
	}
	return body, true
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func parseAnyReply(body []byte) string {
	var openaiP struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &openaiP); err == nil && len(openaiP.Choices) > 0 && openaiP.Choices[0].Message.Content != "" {
		return openaiP.Choices[0].Message.Content
	}
	var anthP struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &anthP); err == nil && len(anthP.Content) > 0 && anthP.Content[0].Text != "" {
		return anthP.Content[0].Text
	}
	var cfP struct {
		Result struct {
			Response string `json:"response"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &cfP); err == nil && cfP.Result.Response != "" {
		return cfP.Result.Response
	}
	return ""
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Message string    `json:"message"`
		History []chatMsg `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Message) == "" {
		jsonOut(w, 400, map[string]string{"error": `need {"message":"..."}`})
		return
	}
	msgs := append([]chatMsg{{"system", "You are a concise, friendly assistant inside seanneol's homelab lab page."}}, in.History...)
	msgs = append(msgs, chatMsg{"user", in.Message})

	provider := strings.ToLower(strings.TrimSpace(os.Getenv("CHAT_PROVIDER")))
	if provider == "" {
		provider = "arena"
	}
	var body []byte
	var code int
	var err error

	switch provider {
	case "groq":
		key := os.Getenv("GROQ_API_KEY")
		if key == "" {
			jsonOut(w, 500, map[string]string{"error": "groq not wired — .env needs GROQ_API_KEY"})
			return
		}
		model := os.Getenv("CHAT_MODEL")
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		payload, _ := json.Marshal(map[string]any{"model": model, "messages": msgs})
		body, code, err = httpPostJSON("https://api.groq.com/openai/v1/chat/completions", "Bearer "+key, nil, payload, 90)
	case "arena", "arena-anthropic":
		key := os.Getenv("ARENA_API_KEY")
		if key == "" {
			jsonOut(w, 500, map[string]string{"error": "arena not wired — .env needs ARENA_API_KEY"})
			return
		}
		base := os.Getenv("ARENA_BASE_URL")
		if base == "" {
			base = "https://api.preview.arena.ai"
		}
		base = strings.TrimRight(base, "/")
		model := os.Getenv("CHAT_MODEL")
		if model == "" {
			model = "coding-router-preview"
		}
		if provider == "arena-anthropic" {
			url := base + "/v1/messages"
			payload, _ := json.Marshal(map[string]any{
				"model":      model,
				"max_tokens": 1024,
				"system":     msgs[0].Content,
				"messages":   msgs[1:],
			})
			body, code, err = httpPostJSON(url, "", map[string]string{
				"x-api-key":         key,
				"anthropic-version": "2023-06-01",
			}, payload, 90)
		} else {
			url := base + "/v1/chat/completions"
			payload, _ := json.Marshal(map[string]any{"model": model, "messages": msgs})
			body, code, err = httpPostJSON(url, "Bearer "+key, nil, payload, 90)
		}
	case "cloudflare", "cf", "workersai":
		payload, _ := json.Marshal(map[string]any{"messages": msgs})
		var ok bool
		body, ok = cfRun(w, "@cf/meta/llama-3.1-8b-instruct", payload)
		if !ok {
			return
		}
		code = 200
	default:
		jsonOut(w, 400, map[string]string{"error": "unknown CHAT_PROVIDER: " + provider})
		return
	}

	if err != nil {
		jsonOut(w, 502, map[string]string{"error": "upstream unreachable: " + err.Error()})
		return
	}
	if code != 200 {
		jsonOut(w, 502, map[string]string{"error": fmt.Sprintf("%s brain said %d: %s", provider, code, string(body))})
		return
	}
	if reply := parseAnyReply(body); reply != "" {
		jsonOut(w, 200, map[string]string{"reply": reply, "provider": provider})
		return
	}
	jsonOut(w, 502, map[string]string{"error": "unparseable brain answer: " + string(body)})
}

func imagineHandler(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Prompt string `json:"prompt"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Prompt) == "" {
		jsonOut(w, 400, map[string]string{"error": `need {"prompt":"..."}`})
		return
	}
	if in.Width == 0 || in.Height == 0 {
		in.Width, in.Height = 1024, 1024
	}
	payload, _ := json.Marshal(map[string]any{
		"prompt": in.Prompt, "steps": 4,
	})
	body, ok := cfRun(w, "@cf/black-forest-labs/flux-1-schnell", payload)
	if !ok {
		return
	}
	var final []byte
	if len(body) > 0 && body[0] == '{' {
		var p struct {
			Result json.RawMessage `json:"result"`
		}
		json.Unmarshal(body, &p)
		var imgs struct {
			Image string `json:"image"`
		}
		if err := json.Unmarshal(p.Result, &imgs); err != nil || imgs.Image == "" {
			jsonOut(w, 502, map[string]string{"error": "unparseable imagine answer: " + string(body)})
			return
		}
		s := imgs.Image
		if strings.HasPrefix(s, "data:") {
			if i := strings.Index(s, ","); i >= 0 {
				s = s[i+1:]
			}
		}
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			jsonOut(w, 502, map[string]string{"error": "bad base64 from brain: " + err.Error()})
			return
		}
		final = b
	} else {
		final = body
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="imagine.png"`)
	w.Write(final)
}

func resizeNN(img image.Image, w int) image.Image {
	b := img.Bounds()
	ow, oh := b.Dx(), b.Dy()
	if ow == 0 || w >= ow {
		return img
	}
	h := oh * w / ow
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, img.At(b.Min.X+x*ow/w, b.Min.Y+y*oh/h))
		}
	}
	return dst
}

func imgConvert(data []byte, format string, width int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("not a decodable image (png/jpg/gif): %v", err)
	}
	if width > 0 {
		img = resizeNN(img, width)
	}
	buf := new(bytes.Buffer)
	switch format {
	case "png":
		err = png.Encode(buf, img)
	case "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	case "gif":
		err = gif.Encode(buf, img, nil)
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonOut(w, 400, map[string]string{"error": "bad upload: " + err.Error()})
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		jsonOut(w, 400, map[string]string{"error": "no file in field 'file'"})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		jsonOut(w, 500, map[string]string{"error": err.Error()})
		return
	}
	mode := r.FormValue("mode")
	width, _ := strconv.Atoi(r.FormValue("width"))
	base := strings.TrimSuffix(hdr.Filename, path.Ext(hdr.Filename))
	if base == "" {
		base = "converted"
	}
	var out []byte
	ext := "bin"
	ctype := "application/octet-stream"
	switch mode {
	case "img2png", "img2jpg", "img2gif":
		format := strings.TrimPrefix(mode, "img2")
		out, err = imgConvert(data, format, width)
		ext = format
		ctype = "image/" + format
	case "resize":
		format := "png"
		lower := strings.ToLower(hdr.Filename)
		switch {
		case strings.HasSuffix(lower, ".png"):
			format = "png"
		case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
			format = "jpg"
		case strings.HasSuffix(lower, ".gif"):
			format = "gif"
		}
		out, err = imgConvert(data, format, width)
		ext = format
		ctype = "image/" + format
	case "base64-enc":
		out = []byte(base64.StdEncoding.EncodeToString(data))
		ext, ctype = "txt", "text/plain"
	case "base64-dec":
		var buf []byte
		buf, err = base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		out = buf
		ext, ctype = "txt", "application/octet-stream"
	case "csv2json":
		out, err = csv2json(data)
		ext, ctype = "json", "application/json"
	case "json2csv":
		out, err = json2csv(data)
		ext, ctype = "csv", "text/csv"
	case "md2html":
		out = md2html(data)
		ext, ctype = "html", "text/html"
	case "json-pretty":
		var v any
		if err = json.Unmarshal(data, &v); err == nil {
			buf := new(bytes.Buffer)
			en := json.NewEncoder(buf)
			en.SetIndent("", "  ")
			err = en.Encode(v)
			out = buf.Bytes()
		}
		ext, ctype = "json", "application/json"
	case "json-minify":
		var v any
		if err = json.Unmarshal(data, &v); err == nil {
			out, err = json.Marshal(v)
		}
		ext, ctype = "json", "application/json"
	default:
		jsonOut(w, 400, map[string]string{"error": "unknown mode: " + mode})
		return
	}
	if err != nil {
		jsonOut(w, 422, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.%s"`, base, ext))
	w.Write(out)
}

func csv2json(data []byte) ([]byte, error) {
	recs, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return []byte("[]"), nil
	}
	head := recs[0]
	rows := []map[string]string{}
	for _, rec := range recs[1:] {
		m := map[string]string{}
		for i, cell := range rec {
			if i < len(head) {
				m[head[i]] = cell
			}
		}
		rows = append(rows, m)
	}
	buf := new(bytes.Buffer)
	en := json.NewEncoder(buf)
	en.SetIndent("", "  ")
	err = en.Encode(rows)
	return buf.Bytes(), err
}

func json2csv(data []byte) ([]byte, error) {
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, fmt.Errorf("expected a JSON array of objects: %v", err)
	}
	if len(arr) == 0 {
		return []byte{}, nil
	}
	keyset := map[string]bool{}
	var keys []string
	for _, row := range arr {
		for k := range row {
			if !keyset[k] {
				keyset[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.Write(keys)
	for _, row := range arr {
		rec := make([]string, len(keys))
		for i, k := range keys {
			rec[i] = fmt.Sprint(row[k])
		}
		w.Write(rec)
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func pair(s, mark, open, close string) string {
	for {
		i := strings.Index(s, mark)
		if i < 0 {
			return s
		}
		j := strings.Index(s[i+len(mark):], mark)
		if j < 0 {
			return s
		}
		s = s[:i] + open + s[i+len(mark):i+len(mark)+j] + close + s[i+len(mark)+j+len(mark):]
	}
}

func md2html(data []byte) []byte {
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	inline := func(s string) string {
		s = pair(s, "**", "<strong>", "</strong>")
		s = pair(s, "`", "<code>", "</code>")
		return s
	}
	var out strings.Builder
	out.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body>\n")
	inUL := false
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "- "):
			if !inUL {
				out.WriteString("<ul>")
				inUL = true
			}
			out.WriteString("<li>" + inline(esc.Replace(t[2:])) + "</li>")
		default:
			if inUL {
				out.WriteString("</ul>")
				inUL = false
			}
			switch {
			case strings.HasPrefix(t, "### "):
				out.WriteString("<h3>" + inline(esc.Replace(t[4:])) + "</h3>")
			case strings.HasPrefix(t, "## "):
				out.WriteString("<h2>" + inline(esc.Replace(t[3:])) + "</h2>")
			case strings.HasPrefix(t, "# "):
				out.WriteString("<h1>" + inline(esc.Replace(t[2:])) + "</h1>")
			case t == "":
			default:
				out.WriteString("<p>" + inline(esc.Replace(t)) + "</p>")
			}
		}
	}
	if inUL {
		out.WriteString("</ul>")
	}
	out.WriteString("\n</body></html>")
	return []byte(out.String())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /api/convert", convertHandler)
	mux.HandleFunc("POST /api/chat", chatHandler)
	mux.HandleFunc("POST /api/imagine", imagineHandler)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	log.Println("lab listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
