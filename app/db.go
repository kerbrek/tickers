package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func connectLoop(driver string, DSN string, timeout time.Duration) (*sql.DB, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("connectLoop: db connection failed after %s timeout", timeout)

		case <-ticker.C:
			database, _ := sql.Open("postgres", DSN)
			err := database.Ping()
			if err == nil {
				return database, nil
			}
		}
	}
}

func makeDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%s sslmode=%s",
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_SSLMODE"),
	)

	database, err := connectLoop("postgres", connStr, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("makeDB: %w", err)
	}

	db = database
	return database, nil
}

func createTables() error {
	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS tickers (
			symbol varchar(20) PRIMARY KEY,
			price_24h float NOT NULL,
			volume_24h float NOT NULL,
			last_trade_price float NOT NULL
		)`,
	)
	if err != nil {
		return fmt.Errorf("createTables: %w", err)
	}

	return nil
}

func dropTables() error {
	_, err := db.Exec("DROP TABLE IF EXISTS tickers")
	if err != nil {
		return fmt.Errorf("dropTables: %w", err)
	}

	return nil
}

func saveTickers(tickers []TickerIn) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("saveTickers: %w", err)
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
		return fmt.Errorf("saveTickers: %w", err)
	}
	defer stmt.Close()

	for _, t := range tickers {
		_, err := stmt.Exec(t.Symbol, t.Price24h, t.Volume24h, t.LastTradePrice)
		if err != nil {
			return fmt.Errorf("saveTickers: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("saveTickers: %w", err)
	}

	return nil
}

func readTickers() (map[string]TickerOut, error) {
	rows, err := db.Query("SELECT symbol, price_24h, volume_24h, last_trade_price FROM tickers ORDER BY symbol")
	if err != nil {
		return nil, fmt.Errorf("readTickers: %w", err)
	}
	defer rows.Close()

	tickers := make(map[string]TickerOut)

	for rows.Next() {
		var symbol string
		var t TickerOut
		if err := rows.Scan(&symbol, &t.Price, &t.Volume, &t.LastTrade); err != nil {
			return nil, fmt.Errorf("readTickers: %w", err)
		}
		tickers[symbol] = t
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("readTickers: %w", err)
	}

	return tickers, nil
}
