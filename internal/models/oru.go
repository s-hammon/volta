package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type ORU struct {
	MSH MessageModel  `json:"MSH"`
	PID PatientModel  `json:"PID"`
	PV1 VisitModel    `json:"PV1"`
	ORC OrderModel    `json:"ORC"`
	OBR ExamModel     `json:"OBR"`
	OBX []ReportModel `json:"OBX"`
}

func NewORU(msgMap map[string]interface{}) (ORU, error) {
	b, err := json.Marshal(msgMap)
	if err != nil {
		return ORU{}, err
	}

	oru := ORU{}
	if err = json.Unmarshal(b, &oru); err != nil {
		return ORU{}, err
	}

	return oru, nil
}

func (oru *ORU) ToDB(ctx context.Context, db *database.Queries) (Response, error) {
	var r Response
	entities := map[string]interface{}{}

	p := oru.PID.ToEntity()
	m := oru.MSH.ToEntity()
	v := oru.PV1.ToEntity(m.SendingFac, oru.PID.MRN)
	o := oru.ORC.ToEntity()
	e := oru.OBR.ToEntity(v.Site.Code, o.CurrentStatus, oru.PID.MRN)

	site, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating site: "+err.Error(), r, entities)
	}
	entities["site"] = site

	patient, err := p.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating patient: "+err.Error(), r, entities)
	}
	entities["patient"] = patient

	mrn, err := v.MRN.ToDB(ctx, site.ID, patient.ID, db)
	if err != nil {
		return handleError("error creating MRN: "+err.Error(), r, entities)
	}
	entities["mrn"] = mrn

	visit, err := v.ToDB(ctx, site.ID, mrn.ID, db)
	if err != nil {
		return handleError("error creating visit: "+err.Error(), r, entities)
	}
	entities["visit"] = visit

	physician, err := o.Provider.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating physician: "+err.Error(), r, entities)
	}
	entities["physician"] = physician

	order, err := o.ToDB(ctx, site.ID, visit.ID, mrn.ID, physician.ID, db)
	if err != nil {
		return handleError("error creating order: "+err.Error(), r, entities)
	}
	entities["order"] = order

	exam, err := e.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating exam: "+err.Error(), r, entities)
	}

	report := oru.getReport(exam, physician)
}

func (oru *ORU) getReport(exam entity.Exam, radiologist entity.Physician) entity.Report {
	body := ""

	for _, obx := range oru.OBX {
		body += obx.ObservationValue + "\n"
	}

	submitDT, err := time.Parse("20060102150405", oru.OBR.ObservationDT)
	if err != nil {
		submitDT = time.Now()
	}

	return entity.Report{
		Exam:        exam,
		Radiologist: radiologist,
		Body:        body,
		Impression:  oru.OBX[0].ObservationSubID,
		Status:      objects.NewReportStatus(oru.OBR.Status),
		SubmittedDT: submitDT,
	}
}
