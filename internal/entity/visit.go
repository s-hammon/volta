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

func DBtoVisit(visit database.GetVisitBySiteIdNumberRow) Visit {
	return Visit{
		Base: Base{
			ID:        int(visit.ID),
			CreatedAt: visit.CreatedAt.Time,
			UpdatedAt: visit.UpdatedAt.Time,
		},
		VisitNo: visit.Number,
		Site: Site{
			Base: Base{
				ID:        int(visit.SiteID.Int32),
				CreatedAt: visit.SiteCreatedAt.Time,
				UpdatedAt: visit.SiteUpdatedAt.Time,
			},
			Code:    visit.SiteCode.String,
			Name:    visit.SiteName.String,
			Address: visit.SiteAddress.String,
		},
		Type: objects.PatientType(visit.PatientType),
	}
}

func (v *Visit) ToDB(ctx context.Context, siteID int32, mrnID int64, db *database.Queries) (int64, error) {
	// TODO: if v.VisitNo == "", use the accession
	visit, err := db.CreateVisit(ctx, database.CreateVisitParams{
		SiteID:      pgtype.Int4{Int32: siteID, Valid: true},
		MrnID:       pgtype.Int8{Int64: mrnID, Valid: true},
		Number:      v.VisitNo,
		PatientType: v.Type.Int16(),
	})
	if err != nil {
		return 0, err
	}
	return visit.ID, nil
}
