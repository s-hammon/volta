package models

import (
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type VisitModel struct {
	Facility         string `hl7:"MSH.4"`
	MRN              CX     `hl7:"PID.3"`
	VisitNo          string `hl7:"PV1.19"`
	Class            string `hl7:"PV1.2"`
	AssignedLocation PL     `hl7:"PV1.3"`
}

func (v *VisitModel) ToEntity() entity.Visit {
	visit := entity.Visit{
		VisitNo: v.VisitNo,
		Site:    entity.Site{Code: v.Facility},
		MRN: entity.MRN{
			Value:              v.MRN.ID,
			AssigningAuthority: v.MRN.AssigningAuthority,
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
