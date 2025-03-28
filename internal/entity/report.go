package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Report struct {
	Base
	Radiologist Physician
	Body        string
	Impression  string
	Status      objects.ReportStatus
	SubmittedDT time.Time
}

func (r *Report) ToDB(ctx context.Context, db *database.Queries) (database.Report, error) {
	var submitDT pgtype.Timestamp
	if err := submitDT.Scan(r.SubmittedDT); err != nil {
		return database.Report{}, err
	}

	res, err := db.CreateReport(ctx, database.CreateReportParams{
		RadiologistID: pgtype.Int8{Int64: int64(r.Radiologist.ID), Valid: true},
		Body:          r.Body,
		Impression:    r.Impression,
		ReportStatus:  r.Status.String(),
		SubmittedDt:   submitDT,
	})
	if err != nil {
		return database.Report{}, err
	}

	return res, nil
}
