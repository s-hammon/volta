package entity

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Visit struct {
	Base
	VisitNo string
	Site    Site
	MRN     MRN
	Type    objects.PatientType
}

func (v *Visit) ToDB(ctx context.Context, siteID int32, mrnID int64, db *database.Queries) (database.Visit, error) {
	var mID pgtype.Int8
	if err := mID.Scan(mrnID); err != nil {
		return database.Visit{}, err
	}

	res, err := db.CreateVisit(ctx, database.CreateVisitParams{
		SiteID:      pgtype.Int4{Int32: siteID, Valid: true},
		MrnID:       mID,
		Number:      v.VisitNo,
		PatientType: v.Type.Int16(),
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetVisitBySiteIdNumber(ctx, database.GetVisitBySiteIdNumberParams{
			SiteID: pgtype.Int4{Int32: siteID, Valid: true},
			Number: v.VisitNo,
		})
		if err == nil {
			return res, nil
		}
	}

	return database.Visit{}, err
}
