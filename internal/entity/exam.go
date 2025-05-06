package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
)

type ExamStatus string

const (
	ExamScheduled  ExamStatus = "SC"
	ExamInProgress ExamStatus = "IP"
	ExamComplete   ExamStatus = "CM"
	ExamCancelled  ExamStatus = "CA"
)

func NewExamStatus(s string) ExamStatus {
	switch s {
	case "SC", "IP", "CM", "CA":
		return ExamStatus(s)
	}
	return ExamScheduled
}

func (o ExamStatus) String() string {
	return string(o)
}

type Exam struct {
	Base
	Accession     string
	MRN           MRN
	Procedure     Procedure
	CurrentStatus ExamStatus
	Provider      Physician
	Site          Site
	Scheduled     time.Time
	Begin         time.Time
	End           time.Time
	Cancelled     time.Time
}

func DBtoExam(exam database.GetExamBySiteIDAccessionRow) Exam {
	return Exam{
		Base: Base{
			ID:        int(exam.ID),
			CreatedAt: exam.CreatedAt.Time,
			UpdatedAt: exam.UpdatedAt.Time,
		},
		Accession: exam.Accession,
		MRN: MRN{
			Base: Base{
				ID:        int(exam.MrnID.Int64),
				CreatedAt: exam.MrnCreatedAt.Time,
				UpdatedAt: exam.MrnUpdatedAt.Time,
			},
			Value: exam.MrnValue.String,
		},
		Procedure: Procedure{
			Base: Base{
				ID:        int(exam.ProcedureID.Int32),
				CreatedAt: exam.ProcedureCreatedAt.Time,
				UpdatedAt: exam.ProcedureUpdatedAt.Time,
			},
			Code:        exam.ProcedureCode.String,
			Description: exam.ProcedureDescription.String,
		},
		Site: Site{
			Base: Base{
				ID:        int(exam.SiteID.Int32),
				CreatedAt: exam.SiteCreatedAt.Time,
				UpdatedAt: exam.SiteUpdatedAt.Time,
			},
			Code:    exam.SiteCode.String,
			Name:    exam.SiteName.String,
			Address: exam.SiteAddress.String,
		},
		Scheduled: exam.ScheduleDt.Time,
		Begin:     exam.BeginExamDt.Time,
		End:       exam.EndExamDt.Time,
	}
}

func (e *Exam) ToDB(ctx context.Context, visitID, mrnID, physID int64, siteID, procedureID int32, db *database.Queries) (int64, error) {
	params := database.CreateExamParams{
		VisitID:             pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:               pgtype.Int8{Int64: int64(mrnID), Valid: true},
		SiteID:              pgtype.Int4{Int32: int32(siteID), Valid: true},
		ProcedureID:         pgtype.Int4{Int32: int32(procedureID), Valid: true},
		OrderingPhysicianID: pgtype.Int8{Int64: physID, Valid: true},
		Accession:           e.Accession,
		CurrentStatus:       e.CurrentStatus.String(),
	}
	e.timestamp(&params)

	fmt.Printf("%+v\n", params)
	exam, err := db.CreateExam(ctx, params)
	if err != nil {
		return 0, err
	}
	return exam.ID, nil
}

func (e *Exam) timestamp(params *database.CreateExamParams) {
	if !e.Scheduled.IsZero() {
		params.ScheduleDt = pgtype.Timestamp{Time: e.Scheduled, Valid: true}
	}
	if !e.Begin.IsZero() {
		params.BeginExamDt = pgtype.Timestamp{Time: e.Begin, Valid: true}
	}
	if !e.End.IsZero() {
		params.EndExamDt = pgtype.Timestamp{Time: e.End, Valid: true}
	}
	if !e.Cancelled.IsZero() {
		params.ExamCancelledDt = pgtype.Timestamp{Time: e.Cancelled, Valid: true}
	}
}
