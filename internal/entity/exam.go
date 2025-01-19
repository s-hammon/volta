package entity

import (
	"time"

	"github.com/google/uuid"
)

type Exam struct {
	ID        uuid.UUID
	Accession string
	Patient   Patient
	Physician Physician
	Procedure Procedure
	Site      Site
	Scheduled time.Time
	Begin     time.Time
	End       time.Time
}
