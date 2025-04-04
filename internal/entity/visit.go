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

func (v *Visit) ToDB(ctx context.Context, siteID int32, mrnID int64, db *database.Queries) (database.Visit, error) {
	// TODO: if v.VisitNo == "", use the accession
	return db.CreateVisit(ctx, database.CreateVisitParams{
		SiteID:      pgtype.Int4{Int32: siteID, Valid: true},
		MrnID:       pgtype.Int8{Int64: mrnID, Valid: true},
		Number:      v.VisitNo,
		PatientType: v.Type.Int16(),
	})
}

func (v *Visit) Equal(other Visit) bool {
	return v.VisitNo == other.VisitNo &&
		v.Site.Equal(other.Site) &&
		v.MRN.Equal(other.MRN) &&
		v.Type == other.Type
}

func (v *Visit) Coalesce(other Visit) {
	if other.VisitNo != "" && v.VisitNo != other.VisitNo {
		v.VisitNo = other.VisitNo
	}
	if other.Type != 0 && v.Type != other.Type {
		v.Type = other.Type
	}
	v.Site.Coalesce(other.Site)
	v.MRN.Coalesce(other.MRN)
}
