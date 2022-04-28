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

func saveTickers(tickers []TickerIn) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO tickers (symbol, price_24h, volume_24h, last_trade_price)
			VALUES ($1, $2, $3, $4) ON CONFLICT (symbol) DO UPDATE
			SET price_24h = excluded.price_24h,
				volume_24h = excluded.volume_24h,
				last_trade_price = excluded.last_trade_price`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickers {
		_, err := stmt.Exec(t.Symbol, t.Price24h, t.Volume24h, t.LastTradePrice)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
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
	rows, err := db.Query("SELECT symbol, price_24h, volume_24h, last_trade_price FROM tickers ORDER BY symbol")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer rows.Close()

	tickersMap := make(map[string]TickerOut)

	for rows.Next() {
		var symbol string
		var t TickerOut
		if err := rows.Scan(&symbol, &t.Price, &t.Volume, &t.LastTrade); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Print(err)
			return
		}
		tickersMap[symbol] = t
	}

	if err := rows.Err(); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	c.IndentedJSON(http.StatusOK, tickersMap)
}
