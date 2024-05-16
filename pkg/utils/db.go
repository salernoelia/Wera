package utils

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

var db *pgxpool.Pool

func init() {
    var err error
    connString := "postgres://myuser:mypassword@localhost:5432/mydatabase"
    db, err = pgxpool.Connect(context.Background(), connString)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }
}

func GetDBConnection() (*pgxpool.Pool, error) {
    return db, nil
}
