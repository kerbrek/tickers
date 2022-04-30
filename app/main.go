package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var db *sql.DB
var tickersURL = "https://api.blockchain.com/v3/exchange/tickers"

type TickerIn struct {
	Symbol         string  `json:"symbol"`
	Price24h       float64 `json:"price_24h"`
	Volume24h      float64 `json:"volume_24h"`
	LastTradePrice float64 `json:"last_trade_price"`
}

type TickerOut struct {
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	LastTrade float64 `json:"last_trade"`
}

func main() {
	db = makeDB()
	createTables()

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	err := updateTickers(&client)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Tickers saved")

	interval := 30 * time.Second
	go updateTickersByInterval(interval, &client)

	addr := fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT"))
	router := setupRouter()
	router.Run(addr)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/tickers", getTickers)
	return router
}

func downloadTickers(client *http.Client) ([]TickerIn, error) {
	resp, err := client.Get(tickersURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf(`%s "%s": %s`, resp.Request.Method, resp.Request.URL, resp.Status)
		return nil, errors.New(errMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tickers := []TickerIn{}
	err = json.Unmarshal(body, &tickers)
	if err != nil {
		return nil, err
	}

	return tickers, nil
}

func updateTickers(client *http.Client) error {
	tickers, err := downloadTickers(client)
	if err != nil {
		return err
	}

	err = saveTickers(tickers)
	if err != nil {
		return err
	}

	return nil
}

func updateTickersByInterval(interval time.Duration, client *http.Client) {
	for range time.Tick(interval) {
		err := updateTickers(client)
		if err != nil {
			log.Print(err)
		} else {
			log.Print("Tickers updated")
		}
	}
}

func getTickers(c *gin.Context) {
	tickers, err := readTickers()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	c.IndentedJSON(http.StatusOK, tickers)
}
