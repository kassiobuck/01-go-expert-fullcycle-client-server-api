package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ExtUSD struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type OutPutBIDUSD struct {
	Bid string
}

func requestUSD() ExtUSD {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		println("New Request error")
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Timeout error")
		panic(err)
	}
	defer res.Body.Close()
	var extUSD ExtUSD
	ioRes, err := io.ReadAll(res.Body)
	err = json.Unmarshal(ioRes, &extUSD)
	if err != nil {
		println("Json unmarshal error")
		panic(err)
	}
	return extUSD
}

func main() {
	// - create server mux
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		extUSD := requestUSD()
		var output OutPutBIDUSD
		output.Bid = extUSD.USDBRL.Bid
		w.Header().Set("Content-Type", "Application-json")
		json.NewEncoder(w).Encode(output)
	})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
