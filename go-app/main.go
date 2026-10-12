package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Note struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// wait for Postgres to be ready (retry on boot)
	for i := 0; i < 15; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Println("waiting for postgres...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("postgres unreachable: ", err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	)`)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/notes", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, text, created_at FROM notes ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		notes := []Note{}
		for rows.Next() {
			var n Note
			rows.Scan(&n.ID, &n.Text, &n.CreatedAt)
			notes = append(notes, n)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notes)
	})
	mux.HandleFunc("POST /api/notes", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
			http.Error(w, `{"error":"need {"text":"..."}"}`, 400)
			return
		}
		_, err := db.Exec("INSERT INTO notes (text) VALUES ($1)", body.Text)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write([]byte(`{"status":"created"}`))
	})

	log.Println("go-app listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}