package entity

import (
	"context"
	"fmt"
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

func (p *Patient) ToDB(ctx context.Context, db *database.Queries) (database.Patient, error) {
	var dob pgtype.Date
	if err := dob.Scan(p.DOB); err != nil {
		return database.Patient{}, err
	}

	var ssn pgtype.Text
	if err := ssn.Scan(p.SSN); err != nil {
		return database.Patient{}, err
	}

	res, err := db.CreatePatient(ctx, database.CreatePatientParams{
		FirstName:  p.Name.First,
		LastName:   p.Name.Last,
		MiddleName: pgtype.Text{String: p.Name.Middle, Valid: true},
		Suffix:     pgtype.Text{String: p.Name.Suffix, Valid: true},
		Prefix:     pgtype.Text{String: p.Name.Prefix, Valid: true},
		Degree:     pgtype.Text{String: p.Name.Degree, Valid: true},
		Dob:        dob,
		Sex:        p.Sex,
		Ssn:        ssn,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetPatientByNameSSN(ctx, database.GetPatientByNameSSNParams{
			FirstName: p.Name.First,
			LastName:  p.Name.Last,
			Dob:       pgtype.Date{Time: p.DOB, Valid: true},
			Ssn:       pgtype.Text{String: p.SSN.String(), Valid: true},
		})
		if err != nil {
			return database.Patient{}, err
		}

		pt := DBtoPatient(res)
		if !pt.Equal(*p) {
			pt.Coalesce(*p)
			res, err := db.UpdatePatient(ctx, database.UpdatePatientParams{
				ID:         int64(pt.ID),
				FirstName:  pt.Name.First,
				LastName:   pt.Name.Last,
				MiddleName: pgtype.Text{String: pt.Name.Middle, Valid: true},
				Suffix:     pgtype.Text{String: pt.Name.Suffix, Valid: true},
				Prefix:     pgtype.Text{String: pt.Name.Prefix, Valid: true},
				Degree:     pgtype.Text{String: pt.Name.Degree, Valid: true},
				Dob:        pgtype.Date{Time: pt.DOB, Valid: true},
				Sex:        pt.Sex,
				Ssn:        pgtype.Text{String: pt.SSN.String(), Valid: true},
			})
			if err != nil {
				return database.Patient{}, err
			}

			return res, nil
		}
	}

	return database.Patient{}, err
}

func (p *Patient) String() string {
	return fmt.Sprintf("Name: %s\tDOB: %v\tSex: %s", p.Name.Record(), p.DOB, p.Sex)
}

func (p *Patient) Equal(other Patient) bool {
	return p.Name.Full() == other.Name.Full() &&
		p.DOB.Equal(other.DOB) &&
		p.Sex == other.Sex &&
		p.SSN == other.SSN
}

func (p *Patient) Coalesce(other Patient) {
	p.Name.Coalesce(other.Name)
	if !other.DOB.IsZero() {
		p.DOB = other.DOB
	}
	if other.Sex != "" {
		p.Sex = other.Sex
	}
	if other.SSN != "" {
		p.SSN = other.SSN
	}
}

type MRN struct {
	Base
	Value string
	// TODO: handle assigning authority
	AssigningAuthority string
}

func (m *MRN) ToDB(ctx context.Context, siteID int32, patientID int64, db *database.Queries) (database.Mrn, error) {
	var ptID pgtype.Int8
	if err := ptID.Scan(patientID); err != nil {
		return database.Mrn{}, err
	}

	res, err := db.CreateMrn(ctx, database.CreateMrnParams{
		SiteID:    siteID,
		PatientID: ptID,
		Mrn:       m.Value,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetMrnBySitePatient(ctx, database.GetMrnBySitePatientParams{
			SiteID:    siteID,
			PatientID: ptID,
		})
		if err == nil {
			return res, nil
		}
	}

	return database.Mrn{}, err
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

type MrnPatientMap map[string]Patient
