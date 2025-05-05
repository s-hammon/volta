package models

import (
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type ExamModel struct {
	FacilityCode     string `hl7:"MSH.4"`
	MRN              CX     `hl7:"PID.3"`
	Accession        string `hl7:"ORC.2"`
	FillerOrderNo    string `hl7:"ORC.3"`
	OrderStatus      string `hl7:"ORC.5"`
	OrderDT          string `hl7:"ORC.9"`
	OrderingProvider XCN    `hl7:"ORC.12"`
	Service          CE     `hl7:"OBR.4"`
	StatusDT         string `hl7:"OBR.22"`
}

func (e *ExamModel) ToEntity() entity.Exam {
	site := entity.Site{Code: e.FacilityCode}
	procedure := entity.Procedure{
		Site:        site,
		Code:        e.Service.Identifier,
		Description: e.Service.Text,
	}
	provider := entity.Physician{
		Name: objects.Name{
			Last:   e.OrderingProvider.FamilyName,
			First:  e.OrderingProvider.GivenName,
			Middle: e.OrderingProvider.MiddleName,
			Suffix: e.OrderingProvider.Suffix,
			Prefix: e.OrderingProvider.Prefix,
			Degree: e.OrderingProvider.Degree,
		},
	}
	exam := entity.Exam{
		Accession: coalesce(e.Accession, e.FillerOrderNo),
		MRN: entity.MRN{
			Value:              e.MRN.ID,
			AssigningAuthority: e.MRN.AssigningAuthority,
		},
		Procedure:     procedure,
		CurrentStatus: entity.NewExamStatus(e.OrderStatus),
		Provider:      provider,
		Site:          site,
	}

	dt := convertCSTtoUTC(e.StatusDT)
	switch exam.CurrentStatus {
	case "SC":
		exam.Scheduled = dt
	case "IP":
		exam.Begin = dt
	case "CM":
		exam.End = dt
	case "CA":
		exam.Cancelled = dt
	}

	return exam
}

func ToEntities(exams []ExamModel) []entity.Exam {
	res := make([]entity.Exam, len(exams))
	for i, exam := range exams {
		res[i] = exam.ToEntity()
	}
	return res
}

func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
