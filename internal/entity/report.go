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

func DBtoReport(report database.Report) Report {
	return Report{
		Base: Base{
			ID:        int(report.ID),
			CreatedAt: report.CreatedAt.Time,
			UpdatedAt: report.UpdatedAt.Time,
		},
		Radiologist: Physician{
			Base: Base{ID: int(report.RadiologistID.Int64)},
		},
		Body:        report.Body,
		Impression:  report.Impression,
		Status:      objects.NewReportStatus(report.ReportStatus),
		SubmittedDT: report.SubmittedDt.Time,
	}
}

func (r *Report) ToDB(ctx context.Context, db *database.Queries, radID int64) (int64, error) {
	var submitDT pgtype.Timestamp
	if err := submitDT.Scan(r.SubmittedDT); err != nil {
		return 0, err
	}

	res, err := db.CreateReport(ctx, database.CreateReportParams{
		RadiologistID: pgtype.Int8{Int64: radID, Valid: true},
		Body:          r.Body,
		Impression:    r.Impression,
		ReportStatus:  r.Status.String(),
		SubmittedDt:   submitDT,
	})
	if err != nil {
		return 0, err
	}

	return res.ID, nil
}
