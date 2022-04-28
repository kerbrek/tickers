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
