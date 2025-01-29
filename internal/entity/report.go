package entity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

const ()

type Report struct {
	Base
	Exam        Exam
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
		ExamID:        pgtype.Int8{Int64: int64(r.Exam.ID), Valid: true},
		RadiologistID: pgtype.Int8{Int64: int64(r.Radiologist.ID), Valid: true},
		Body:          r.Body,
		Impression:    r.Impression,
		ReportStatus:  r.Status.String(),
		SubmittedDt:   submitDT,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		// throw error--can only have one report per exam & status
		errMsg := fmt.Sprintf("report already exists for exam %v", r.Exam.ID)
		return database.Report{}, errors.New(errMsg)
	}

	return database.Report{}, err
}
