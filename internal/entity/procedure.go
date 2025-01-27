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

func (p *Procedure) ToDB(ctx context.Context, db *database.Queries) (database.Procedure, error) {
	var sID pgtype.Int4
	if err := sID.Scan(p.Site.ID); err != nil {
		return database.Procedure{}, err
	}

	res, err := db.CreateProcedure(ctx, database.CreateProcedureParams{
		Code:        p.Code,
		Description: p.Description,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetProcedureBySiteIDCode(ctx, database.GetProcedureBySiteIDCodeParams{
			SiteID: sID,
			Code:   p.Code,
		})
		if err == nil {
			return res, nil
		}
	}

	return database.Procedure{}, err
}
