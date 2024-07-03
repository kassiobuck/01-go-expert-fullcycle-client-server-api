package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_DRIVER       = "sqlite3"
	DB_FILE         = "./database.db"
	DB_TIMEOUT      = 10
	REQUEST_TIMEOUT = 200
)

type USDBRL struct {
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
}

type ExtUSD struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type OutPutBIDUSD struct {
	Bid string `json:"bid"`
}

func getUSDBRL() (USDBRL, error) {
	var extUSD ExtUSD
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, REQUEST_TIMEOUT*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return extUSD.USDBRL, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return extUSD.USDBRL, ctx.Err()
		}

		return extUSD.USDBRL, err
	}
	defer res.Body.Close()

	ioRes, err := io.ReadAll(res.Body)
	if err != nil {
		return extUSD.USDBRL, err
	}

	err = json.Unmarshal(ioRes, &extUSD)
	if err != nil {
		return extUSD.USDBRL, err
	}

	return extUSD.USDBRL, err
}

func saveUSDBRL(usdbrl USDBRL) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT*time.Millisecond)
	defer cancel()
	conn, err := sql.Open(DB_DRIVER, DB_FILE)
	if err != nil {
		return err
	}
	defer conn.Close()
	q, err := conn.Prepare("INSERT INTO usdbrl (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, usdbrl.Code, usdbrl.Codein, usdbrl.Name, usdbrl.High, usdbrl.Low, usdbrl.VarBid, usdbrl.PctChange, usdbrl.Bid, usdbrl.Ask, usdbrl.Timestamp, usdbrl.CreateDate)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			ctx.Err()
		}
		return err
	}
	return nil
}

func initDataBase() {
	conn, err := sql.Open(DB_DRIVER, DB_FILE)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS usdbrl ( id INTEGER NOT NULL PRIMARY KEY, code TEXT, codein TEXT, name TEXT, high TEXT, low TEXT, varBid TEXT, pctChange TEXT, bid TEXT, ask TEXT, timestamp TEXT, createDate TEXT)")
	if err != nil {
		panic(err)
	}
}

func main() {
	initDataBase()
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handlerCotacao)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func handlerCotacao(w http.ResponseWriter, r *http.Request) {
	usdbrl, err := getUSDBRL()
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("-> Context error: request timeout")
			log.Println(err)
			w.WriteHeader(http.StatusRequestTimeout)
			return
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	err = saveUSDBRL(usdbrl)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("-> Context error: database timeout")
			log.Println(err)
		} else {
			log.Println(err)
			return
		}
	}
	var output = OutPutBIDUSD{Bid: usdbrl.Bid}
	w.Header().Set("Content-Type", "Application-json")
	json.NewEncoder(w).Encode(output)
}
