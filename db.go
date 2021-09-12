package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(host, port, username, password, dbname, sslmode string) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, dbname, sslmode,
	)

	db, err := sqlx.Connect(
		"postgres",
		connStr,
	)

	if err != nil {
		return nil, err
	}
	return db, nil
}
