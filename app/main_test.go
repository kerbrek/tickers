package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnaeon/go-vcr/v2/recorder"
	"github.com/stretchr/testify/assert"
)

func TestGetTickersRoute(t *testing.T) {
	assert := assert.New(t)

	r, err := recorder.New("fixtures/tickers")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	client := &http.Client{
		Transport: r,
	}

	if _, err := makeDB(); err != nil {
		t.Fatal(err)
	}

	if err := dropTables(); err != nil {
		t.Fatal(err)
	}

	if err := createTables(); err != nil {
		t.Fatal(err)
	}

	tickersIn, err := downloadTickers(client)
	if err != nil {
		t.Fatal(err)
	}

	err = saveTickers(tickersIn)
	if err != nil {
		t.Fatal(err)
	}

	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tickers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(200, w.Code)

	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	tickersOut := make(map[string]TickerOut)
	err = json.Unmarshal(body, &tickersOut)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(len(tickersIn), len(tickersOut))

	for _, tIn := range tickersIn {
		tOut := tickersOut[tIn.Symbol]
		assert.Equal(tIn.Price24h, tOut.Price)
		assert.Equal(tIn.Volume24h, tOut.Volume)
		assert.Equal(tIn.LastTradePrice, tOut.LastTrade)
	}
}
