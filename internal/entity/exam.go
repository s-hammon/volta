package entity

import (
	"time"

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

func (e *Exam) ToDB() (database.Exam, error)
