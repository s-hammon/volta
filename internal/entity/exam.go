package entity

import (
	"time"

	"github.com/google/uuid"
)

type Exam struct {
	ID          uuid.UUID
	Accession   string
	MRN         MRN
	Physician   Physician
	Procedure   Procedure
	Site        Site
	Scheduled   time.Time
	Begin       time.Time
	End         time.Time
	Cancelled   time.Time
	Rescheduled map[time.Time]struct{} // this should be interesting
}
