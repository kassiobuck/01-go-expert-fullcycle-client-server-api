package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

const (
	TXT_FILE        = "cotacao.txt"
	REQUEST_TIMEOUT = 300
)

type ExtBid struct {
	Bid string `json:"bid"`
}

func main() {
	var extBid ExtBid
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, REQUEST_TIMEOUT*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("context error: request timeout")
		}
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		if res.StatusCode == 408 {
			log.Println("Server timeout")
		} else {
			log.Println("Server error")
		}
		return
	}

	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(ioRes, &extBid)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	t := template.Must(template.New("cotacao.txt").Parse("Dólar: {{.Bid}}"))
	err = t.Execute(file, extBid)
	if err != nil {
		panic(err)
	}
}
