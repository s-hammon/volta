package entity

import (
	"github.com/google/uuid"
	"github.com/s-hammon/volta/internal/objects"
)

type Patient struct {
	ID   uuid.UUID
	Name objects.Name
	DOB  string
	SSN  string
}

type MRN struct {
	Value              string
	AssigningAuthority string
}

type MrnPatientMap map[string]Patient
