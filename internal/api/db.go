package api

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/api/models"
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

func (d DB) UpsertORM(ctx context.Context, orm models.ORM) error {
	return orm.ToDB(ctx, d.Queries)
}

func (d DB) InsertORU(ctx context.Context, oru models.ORU) error {
	return oru.ToDB(ctx, d.Queries)
}
