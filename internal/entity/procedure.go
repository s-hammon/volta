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

func (p *Procedure) ToDB(ctx context.Context, siteID int32, db *database.Queries) (database.Procedure, error) {
	return db.CreateProcedure(ctx, database.CreateProcedureParams{
		SiteID:      pgtype.Int4{Int32: siteID, Valid: true},
		Code:        p.Code,
		Description: p.Description,
	})
}

func (p *Procedure) Equal(other Procedure) bool {
	return p.Code == other.Code &&
		p.Description == other.Description &&
		p.Site.Equal(other.Site)
}

func (p *Procedure) Coalesce(other Procedure) {
	if other.Code != "" && p.Code != other.Code {
		p.Code = other.Code
	}
	if other.Description != "" && p.Description != other.Description {
		p.Description = other.Description
	}
	p.Site.Coalesce(other.Site)
}
