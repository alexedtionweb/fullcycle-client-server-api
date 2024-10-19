package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

type QuoteDTO struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	High      string `json:"high"`
	Low       string `json:"low"`
	VarBid    string `json:"varBid"`
	PctChange string `json:"pctChange"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Timestamp string `json:"timestamp"`
}
type QuoteAPIResult = map[string]QuoteDTO

type DatabaseService struct {
	sqlite *sql.DB
}

func (db *DatabaseService) SyncMigrations() {
	_, err := db.sqlite.Exec(`CREATE TABLE
    IF NOT EXISTS currency_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        code TEXT,
        code_in TEXT,
        "name" TEXT,
        high TEXT,
        low TEXT,
        var_bid TEXT,
        pct_change TEXT,
        bid TEXT,
        ask TEXT,
        "timestamp" INTEGER
    )`)

	if err != nil {
		log.Fatal(err)
	}

}

func (db *DatabaseService) SaveCurrencyData(ctx context.Context, q QuoteDTO) error {
	ts, err := strconv.Atoi(q.Timestamp)
	if err != nil {
		ts = time.Now().Nanosecond() / 1000
	}
	_, err = db.sqlite.ExecContext(ctx, `
			INSERT INTO currency_data (
				code, name, high, low, var_bid, pct_change, bid, ask, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		q.Code, q.Name, q.High, q.Low, q.VarBid, q.PctChange, q.Bid, q.Ask, ts,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewDB() *DatabaseService {
	db, err := sql.Open("sqlite", "sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	return &DatabaseService{
		sqlite: db,
	}
}

func fetchQuoteAPI(ctx context.Context) (QuoteAPIResult, error) {
	url := fmt.Sprintf("https://economia.awesomeapi.com.br/json/last/%s-%s", "USD", "BRL")
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quote map[string]QuoteDTO
	if err := json.Unmarshal(data, &quote); err != nil {
		return nil, err
	}
	return quote, nil
}

func getCurrencyQuoteHandlerV1(db *DatabaseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		quote, err := fetchQuoteAPI(reqCtx)

		if err != nil {
			slog.Error("fetchQuoteAPI", "details", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dbCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		if err := db.SaveCurrencyData(dbCtx, quote["USDBRL"]); err != nil {
			slog.Error("db.SaveCurrencyData", "details", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(map[string]string{"bid": quote["USDBRL"].Bid})

		if err != nil {
			slog.Error("fetchQuoteAPI", "details", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

	}

}

func main() {

	db := NewDB()
	db.SyncMigrations()

	SERVER_PORT := ":8080"
	app := http.NewServeMux()

	app.HandleFunc("GET /cotacao", getCurrencyQuoteHandlerV1(db))

	slog.Info("http server started", "port", SERVER_PORT)

	if err := http.ListenAndServe(SERVER_PORT, app); err != nil {
		log.Fatal(err)
	}

}
