package storage

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	db *sql.DB
}

func NewDBStorage(databaseDSN string) *DB {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		log.Printf("failed to prepare DB: %s", err)
		return nil
	}

	return &DB{db: db}
}

func (d *DB) Check(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Set(metric Metric) error {
	return nil
}

func (d *DB) Get(name string) (Metric, error) {
	return Metric{}, nil
}

func (d *DB) GetAll() map[string]Metric {
	return map[string]Metric{}
}

func (d *DB) Close() {
	d.db.Close()
}
