package entity

import (
	"github.com/google/uuid"
	"github.com/s-hammon/volta/internal/objects"
)

type Visit struct {
	ID      uuid.UUID
	VisitNo string
	Site    Site
	MRN     MRN
	Type    objects.PatientType
}
