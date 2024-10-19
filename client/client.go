package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", http.NoBody)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var results struct {
		Bid string `json:"bid"`
	}

	err = json.Unmarshal(body, &results)
	if err != nil {
		log.Fatal(err)
	}

	text := fmt.Sprintf("Dólar: %s\n", results.Bid)

	elapsed := time.Since(start).Milliseconds()

	slog.Info(text, "elapsed", fmt.Sprintf("%vms", elapsed))

	os.WriteFile("cotacao.txt", []byte(text), os.FileMode(0644))
}
