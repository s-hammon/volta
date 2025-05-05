package models

import (
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type ReportModel struct {
	Radiologist      CM_NDL `hl7:"OBR.32"`
	SetID            string `hl7:"OBX.1"`
	ValueType        string `hl7:"OBX.2"`
	Service          CE     `hl7:"OBX.3"`
	ObservationSubID string `hl7:"OBX.4"`
	ObservationValue string `hl7:"OBX.5"`
	ResultStatus     string `hl7:"OBX.11"`
	ObservationDT    string `hl7:"OBX.14"`
}

func GetReport(obx []ReportModel) entity.Report {
	r := entity.Report{}
	for _, o := range obx {
		if o.ObservationValue != "" {
			r.Body += o.ObservationValue + "\n"
		}
	}
	if len(obx) > 0 {
		rad := obx[0].Radiologist.ObservingPractitioner
		r.Radiologist = entity.Physician{
			Name: objects.NewName(
				rad.FamilyName,
				rad.GivenName,
				rad.MiddleName,
				rad.Prefix,
				rad.Suffix,
				rad.Degree,
			),
		}
		r.Impression = obx[0].ObservationValue
		r.SubmittedDT = convertCSTtoUTC(obx[0].ObservationDT)
		r.Status = objects.ReportStatus(obx[0].ResultStatus)
	}
	return r
}
