package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/s-hammon/volta/internal/objects"
)

type Patient struct {
	ID        uuid.UUID
	Name      objects.Name
	DOB       time.Time
	Sex       string
	SSN       string
	HomePhone objects.PhoneNumber
	WorkPhone objects.PhoneNumber
	CellPhone objects.PhoneNumber
}

type MRN struct {
	Value              string
	AssigningAuthority string
}

type MrnPatientMap map[string]Patient
