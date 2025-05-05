package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

var dateFormats = []string{
	"20060102",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"01-02-2006",
	"01/02/06",
}

type PatientModel struct {
	MRN  CX     `hl7:"PID.3"`
	Name XPN    `hl7:"PID.5"`
	DOB  string `hl7:"PID.7"`
	Sex  string `hl7:"PID.8"`
	SSN  string `hl7:"PID.19"`
}

func (p *PatientModel) ToEntity() entity.Patient {
	name := objects.Name{
		Last:   p.Name.LastName,
		First:  p.Name.FirstName,
		Middle: p.Name.MiddleName,
		Suffix: p.Name.Suffix,
		Prefix: p.Name.Prefix,
		Degree: p.Name.Degree,
	}

	return entity.Patient{
		Name: name,
		DOB:  tryParseDOB(p.DOB),
		Sex:  p.Sex,
		SSN:  objects.NewSSN(p.SSN),
	}

}

func tryParseDOB(dob string) time.Time {
	// try to parse dob a few different ways
	// if none work, use current time
	for _, f := range dateFormats {
		dt, err := time.Parse(f, dob)
		if err == nil {
			return dt
		}
	}

	return time.Now()
}
