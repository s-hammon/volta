package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
)

type ExamModel struct {
	Accession     string `json:"OBR.3"`
	Service       CE     `json:"OBR.4"`
	Priority      string `json:"OBR.5"`
	RequestDT     string `json:"OBR.6"`
	ObservationDT string `json:"OBR.7"`
	StatusDT      string `json:"OBR.22"`
}

func (e *ExamModel) ToEntity(siteCode string, status string, mrn CX) entity.Exam {
	site := entity.Site{Code: siteCode}

	procedure := entity.Procedure{
		Site:        site,
		Code:        e.Service.Identifier,
		Description: e.Service.Text,
	}

	exam := entity.Exam{
		Accession: e.Accession,
		MRN: entity.MRN{
			Value:              mrn.ID,
			AssigningAuthority: mrn.AssigningAuthority,
		},
		Procedure: procedure,
		Site:      site,
	}

	dt, err := time.Parse("20060102150405", e.RequestDT)
	if err != nil {
		dt = time.Now()
	}

	switch status {
	case "SC":
		exam.Scheduled = dt
	case "IP":
		exam.Begin = dt
	case "CM":
		exam.End = dt
	case "CA":
		exam.Cancelled = dt
	case "RS":
		exam.Rescheduled[dt] = struct{}{}
	}

	return exam
}
