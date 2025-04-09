package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
)

type ExamModel struct {
	Accession     string `json:"OBR.3"`
	Service       CE     `json:"OBR.4"`
	RequestDT     string `json:"OBR.6"`
	ObservationDT string `json:"OBR.7"` // TODO: change to OBR.22
	StatusDT      string `json:"OBR.22"`
	Status        string `json:"OBR.25"` // TODO: make sure we're using for report type (F - final, A - addendum, P - preliminary)
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

	dt, err := time.Parse("20060102150405", e.StatusDT)
	if err != nil {
		dt = time.Now()
	}

	// TODO: status to uint8 enum
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
