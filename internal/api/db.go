package api

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/database"
)

type DB struct {
	*database.Queries
}

func NewDB(db *pgxpool.Pool) DB {
	return DB{
		Queries: database.New(db),
	}
}
