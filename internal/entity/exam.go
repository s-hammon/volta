package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
)

type Exam struct {
	Base
	Accession   string
	MRN         MRN
	Procedure   Procedure
	Site        Site
	Scheduled   time.Time
	Begin       time.Time
	End         time.Time
	Cancelled   time.Time
	Rescheduled map[time.Time]struct{} // this might be interesting
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
		Scheduled:   exam.ScheduleDt.Time,
		Begin:       exam.BeginExamDt.Time,
		End:         exam.EndExamDt.Time,
		Rescheduled: make(map[time.Time]struct{}),
	}
}

func (e *Exam) ToDB(ctx context.Context, orderID, visitID, mrnID int64, siteID, procedureID int32, currentStatus string, db *database.Queries) (int64, error) {
	params := database.CreateExamParams{
		OrderID:       pgtype.Int8{Int64: orderID, Valid: true},
		VisitID:       pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:         pgtype.Int8{Int64: int64(mrnID), Valid: true},
		SiteID:        pgtype.Int4{Int32: int32(siteID), Valid: true},
		ProcedureID:   pgtype.Int4{Int32: int32(procedureID), Valid: true},
		Accession:     e.Accession,
		CurrentStatus: currentStatus,
	}
	e.timestamp(&params)

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
