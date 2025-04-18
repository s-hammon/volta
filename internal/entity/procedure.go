package entity

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Procedure struct {
	Base
	Site        Site
	Code        string
	Description string
	Specialty   objects.Specialty
	Modality    objects.Modality
}

func (p *Procedure) ToDB(ctx context.Context, siteID int32, db *database.Queries) (int32, error) {
	procedure, err := db.CreateProcedure(ctx, database.CreateProcedureParams{
		SiteID:      pgtype.Int4{Int32: siteID, Valid: true},
		Code:        p.Code,
		Description: p.Description,
	})
	if err != nil {
		return 0, err
	}
	return procedure.ID, nil
}
