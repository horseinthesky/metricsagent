package server

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func (s *Server) runDB(ctx context.Context) {
	db, err := sql.Open("pgx", s.config.DatabaseDSN)
	if err != nil {
		log.Printf("failed to open DB connection: %s", err)
		return
	}

	s.db = db

	<-ctx.Done()

	s.db.Close()
	log.Println("DB connection closed")
}
