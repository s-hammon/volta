package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func (p *Patient) ToDB(ctx context.Context, db *database.Queries) (int64, error) {
	patient, err := db.CreatePatient(ctx, database.CreatePatientParams{
		FirstName:  p.Name.First,
		LastName:   p.Name.Last,
		MiddleName: pgtype.Text{String: p.Name.Middle, Valid: true},
		Suffix:     pgtype.Text{String: p.Name.Suffix, Valid: true},
		Prefix:     pgtype.Text{String: p.Name.Prefix, Valid: true},
		Degree:     pgtype.Text{String: p.Name.Degree, Valid: true},
		Dob:        pgtype.Date{Time: p.DOB, Valid: true},
		Sex:        p.Sex,
		Ssn:        pgtype.Text{String: p.SSN.String(), Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return patient.ID, nil
}

type MRN struct {
	Base
	Value string
	// TODO: handle assigning authority
	AssigningAuthority string
}

func (m *MRN) ToDB(ctx context.Context, siteID int32, patientID int64, db *database.Queries) (int64, error) {
	mrn, err := db.CreateMrn(ctx, database.CreateMrnParams{
		SiteID:    siteID,
		PatientID: pgtype.Int8{Int64: patientID, Valid: true},
		Mrn:       m.Value,
	})
	if err != nil {
		return 0, err
	}
	return mrn.ID, nil
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
