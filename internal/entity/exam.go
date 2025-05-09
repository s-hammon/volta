package entity

import (
	"time"

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

func (e *Exam) timestamp(params *database.CreateExamParams) {
	if !e.Scheduled.IsZero() {
		params.ScheduleDt.Time = e.Scheduled
	}
	if !e.Begin.IsZero() {
		params.BeginExamDt.Time = e.Begin
	}
	if !e.End.IsZero() {
		params.EndExamDt.Time = e.End
	}
	if !e.Cancelled.IsZero() {
		params.ExamCancelledDt.Time = e.Cancelled
	}
}
