package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func makeDB() *sql.DB {
	connStr := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%s sslmode=%s",
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"),
	)

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := database.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	return database
}

func createTables() {
	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS tickers (
			symbol varchar(20) PRIMARY KEY,
			price_24h float NOT NULL,
			volume_24h float NOT NULL,
			last_trade_price float NOT NULL
		)`,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func dropTables() {
	_, err := db.Exec("DROP TABLE IF EXISTS tickers")
	if err != nil {
		log.Fatal(err)
	}
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

func readTickers() (map[string]TickerOut, error) {
	rows, err := db.Query("SELECT symbol, price_24h, volume_24h, last_trade_price FROM tickers ORDER BY symbol")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickers := make(map[string]TickerOut)

	for rows.Next() {
		var symbol string
		var t TickerOut
		if err := rows.Scan(&symbol, &t.Price, &t.Volume, &t.LastTrade); err != nil {
			return nil, err
		}
		tickers[symbol] = t
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickers, nil
}
