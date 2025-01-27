package entity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
)

type Exam struct {
	Base
	Accession   string
	MRN         MRN
	Physician   Physician
	Procedure   Procedure
	Site        Site
	Priority    string
	Scheduled   time.Time
	Begin       time.Time
	End         time.Time
	Cancelled   time.Time
	Rescheduled map[time.Time]struct{} // this should be interesting
}

func (e *Exam) ToDB(ctx context.Context, orderID, visitID, mrnID int64, siteID, procedureID int32, currentStatus string, db *database.Queries) (database.Exam, error) {
	fmt.Printf("site_id: %v\n", siteID)
	res, err := db.CreateExam(ctx, database.CreateExamParams{
		OrderID:       pgtype.Int8{Int64: orderID, Valid: true},
		VisitID:       pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:         pgtype.Int8{Int64: int64(mrnID), Valid: true},
		SiteID:        pgtype.Int4{Int32: int32(siteID), Valid: true},
		ProcedureID:   pgtype.Int4{Int32: int32(procedureID), Valid: true},
		Accession:     e.Accession,
		CurrentStatus: currentStatus,
		ScheduleDt:    pgtype.Timestamp{Time: e.Scheduled, Valid: true},
		BeginExamDt:   pgtype.Timestamp{Time: e.Begin, Valid: true},
		EndExamDt:     pgtype.Timestamp{Time: e.End, Valid: true},
	})
	if err != nil {
		return database.Exam{}, err
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetExamBySiteIDAccession(ctx, database.GetExamBySiteIDAccessionParams{
			SiteID:    pgtype.Int4{Int32: int32(siteID), Valid: true},
			Accession: e.Accession,
		})
		if err != nil {
			return database.Exam{}, err
		}
	}

	updated, err := db.UpdateExamByID(ctx, database.UpdateExamByIDParams{
		ID:            res.ID,
		OrderID:       pgtype.Int8{Int64: orderID, Valid: true},
		VisitID:       pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:         pgtype.Int8{Int64: int64(mrnID), Valid: true},
		SiteID:        pgtype.Int4{Int32: int32(siteID), Valid: true},
		ProcedureID:   pgtype.Int4{Int32: int32(procedureID), Valid: true},
		Accession:     e.Accession,
		CurrentStatus: currentStatus,
		ScheduleDt:    res.ScheduleDt,
		BeginExamDt:   res.BeginExamDt,
		EndExamDt:     res.EndExamDt,
	})
	if err != nil {
		return database.Exam{}, errors.New("error updating exam: " + err.Error())
	}

	return updated, nil
}
