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
			// TODO: make this a little SAFER
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
		vsDB, err := db.GetVisitBySiteIdNumber(ctx, database.GetVisitBySiteIdNumberParams{
			SiteID: pgtype.Int4{Int32: siteID, Valid: true},
			Number: v.VisitNo,
		})
		if err != nil {
			return database.Visit{}, err
		}

		vs := DBtoVisit(vsDB)
		if !vs.Equal(*v) {
			vs.Coalesce(*v)
			siteID := pgtype.Int4{}
			if err = siteID.Scan(vs.Site.ID); err != nil {
				return database.Visit{}, err
			}
			return db.UpdateVisit(ctx, database.UpdateVisitParams{
				ID:          int64(vs.ID),
				SiteID:      siteID,
				MrnID:       pgtype.Int8{Int64: int64(vs.MRN.ID), Valid: true},
				Number:      vs.VisitNo,
				PatientType: vs.Type.Int16(),
			})
		}
	}

	return database.Visit{}, err
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
