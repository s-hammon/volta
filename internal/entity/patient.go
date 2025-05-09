package entity

import (
	"time"

	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Patient struct {
	Base
	Name objects.Name
	DOB  time.Time
	Sex  string
	// TODO: encrypt SSN
	SSN       objects.SSN
	HomePhone objects.PhoneNumber
	WorkPhone objects.PhoneNumber
	CellPhone objects.PhoneNumber
}

func DBtoPatient(patient database.Patient) Patient {
	// TODO: deal w/ phone numbers
	return Patient{
		Base: Base{
			ID:        int(patient.ID),
			CreatedAt: patient.CreatedAt.Time,
			UpdatedAt: patient.UpdatedAt.Time,
		},
		Name: objects.Name{
			First:  patient.FirstName,
			Last:   patient.LastName,
			Middle: patient.MiddleName.String,
			Suffix: patient.Suffix.String,
			Prefix: patient.Prefix.String,
			Degree: patient.Degree.String,
		},
		DOB: patient.Dob.Time,
		Sex: patient.Sex,
		SSN: objects.SSN(patient.Ssn.String),
	}
}

type MRN struct {
	Base
	Value string
	// TODO: handle assigning authority
	AssigningAuthority string
}

func (m *MRN) Equal(other MRN) bool {
	return m.Value == other.Value &&
		m.AssigningAuthority == other.AssigningAuthority
}

func (m *MRN) Coalesce(other MRN) {
	if other.Value != "" && m.Value != other.Value {
		m.Value = other.Value
	}
	if other.AssigningAuthority != "" && m.AssigningAuthority != other.AssigningAuthority {
		m.AssigningAuthority = other.AssigningAuthority
	}
}
