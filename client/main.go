package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type ExtBid struct {
	Bid string
}

func main() {
	var extBid ExtBid
	req, err := http.Get("http://localhost:8080/cotacao")
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	res, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(res, &extBid)
	if err != nil {
		panic(err)
	}

	println(string(extBid.Bid))
}
