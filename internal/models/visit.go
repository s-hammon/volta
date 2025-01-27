package models

import (
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type VisitModel struct {
	VisitNo          string `json:"PV1.19"`
	Class            string `json:"PV1.2"`
	AssignedLocation PL     `json:"PV1.3"`
}

func (v *VisitModel) ToEntity(siteCode string, mrn CX) entity.Visit {
	site := entity.Site{Code: siteCode}

	visit := entity.Visit{
		VisitNo: v.VisitNo,
		Site:    site,
		MRN: entity.MRN{
			Value:              mrn.ID,
			AssigningAuthority: mrn.AssigningAuthority,
		},
	}

	switch v.Class {
	case "I":
		visit.Type = objects.InPatient
	case "E":
		visit.Type = objects.EdPatient
	default:
		visit.Type = objects.OutPatient
	}

	return visit
}
