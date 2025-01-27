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
	Name      objects.Name
	DOB       time.Time
	Sex       string
	SSN       string
	HomePhone objects.PhoneNumber
	WorkPhone objects.PhoneNumber
	CellPhone objects.PhoneNumber
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
			Ssn:       pgtype.Text{String: p.SSN, Valid: true},
		})
		if err == nil {
			return res, nil
		}
	}

	return database.Patient{}, err
}

func (p *Patient) String() string {
	return fmt.Sprintf("Name: %s\tDOB: %v\tSex: %s", p.Name.Record(), p.DOB, p.Sex)
}

type MRN struct {
	Base
	Value              string
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

type MrnPatientMap map[string]Patient
