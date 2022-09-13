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

func (d *DB) Init(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS metrics (
			id text PRIMARY KEY,
			mtype text NOT NULL,
			delta bigint,
			value double precision
		)
	`

	if _, err := d.db.ExecContext(ctx, schema); err != nil {
		return err
	}

	log.Println("postgresql database initialized")

	return nil
}

func (d *DB) Check(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Set(metric Metric) error {
	var err error

	switch metric.MType {
	case Counter.String():
		_, err = d.db.Exec(`
			INSERT INTO metrics(id, mtype, delta) VALUES($1,$2,$3)
			 ON CONFLICT (id) DO UPDATE
			 SET mtype = $2, delta = metrics.delta + $3
		`, metric.ID, metric.MType, metric.Delta)
		if err != nil {
			return err
		}
	case Gauge.String():
		_, err = d.db.Exec(`
			INSERT INTO metrics(id, mtype, value) VALUES($1,$2,$3)
			 ON CONFLICT (id) DO UPDATE
			 SET mtype = $2, value = $3
		`, metric.ID, metric.MType, metric.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) SetBulk(metrics []Metric) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var stmt *sql.Stmt

	for _, metric := range metrics {
		switch metric.MType {
		case Counter.String():
			stmt, err = tx.Prepare(`
				INSERT INTO metrics(id, mtype, delta) VALUES($1,$2,$3)
				ON CONFLICT (id) DO UPDATE
				SET mtype = $2, delta = metrics.delta + $3
			`)
			if err != nil {
				return err
			}
			if _, err = stmt.Exec(metric.ID, metric.MType, metric.Delta); err != nil {
				return err
			}
		case Gauge.String():
			stmt, err = tx.Prepare(`
				INSERT INTO metrics(id, mtype, value) VALUES($1,$2,$3)
				ON CONFLICT (id) DO UPDATE
				SET mtype = $2, value = $3
			`)
			if err != nil {
				return err
			}
			if _, err = stmt.Exec(metric.ID, metric.MType, metric.Value); err != nil {
				return err
			}
		}
	}
	defer stmt.Close()

	return tx.Commit()
}
func (d *DB) Get(ctx context.Context, name string) (Metric, error) {
	query := `SELECT id, mtype, delta, value FROM metrics WHERE id=$1`

	metric := Metric{}
	if err := d.db.QueryRowContext(ctx, query, name).Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		log.Printf("failed to query db: %s", err)
		return Metric{}, err
	}

	return metric, nil
}

func (d *DB) GetAll(ctx context.Context) (map[string]Metric, error) {
	query := `SELECT * FROM metrics`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recs := make([]Metric, 0)

	for rows.Next() {
		var rec Metric

		err = rows.Scan(&rec.ID, &rec.MType, &rec.Delta, &rec.Value)
		if err != nil {
			return nil, err
		}

		recs = append(recs, rec)

		err = rows.Err()
		if err != nil {
			return nil, err
		}
	}

	newDB := map[string]Metric{}
	for _, i := range recs {
		newDB[i.ID] = i
	}

	return newDB, nil
}

func (d *DB) Close() {
	d.db.Close()
}
