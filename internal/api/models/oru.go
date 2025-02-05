package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type ORU struct {
	MSH MessageModel  `hl7:"MSH"`
	PID PatientModel  `hl7:"PID"`
	PV1 VisitModel    `hl7:"PV1"`
	ORC OrderModel    `hl7:"ORC"`
	OBR ExamModel     `hl7:"OBR"`
	OBX []ReportModel `hl7:"OBX"`
}

func (oru *ORU) ToDB(ctx context.Context, db *database.Queries) error {
	p := oru.PID.ToEntity()
	m := oru.MSH.ToEntity()
	v := oru.PV1.ToEntity(m.SendingFac, oru.PID.MRN)
	o := oru.ORC.ToEntity()
	e := oru.OBR.ToEntity(v.Site.Code, o.CurrentStatus, oru.PID.MRN)

	site, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating site: " + err.Error())
	}

	patient, err := p.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating patient: " + err.Error())
	}

	mrn, err := v.MRN.ToDB(ctx, site.ID, patient.ID, db)
	if err != nil {
		return errors.New("error creating mrn: " + err.Error())
	}

	visit, err := v.ToDB(ctx, site.ID, mrn.ID, db)
	if err != nil {
		return errors.New("error creating visit: " + err.Error())
	}

	physician, err := o.Provider.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating physician: " + err.Error())
	}

	order, err := o.ToDB(ctx, site.ID, visit.ID, mrn.ID, physician.ID, db)
	if err != nil {
		return errors.New("error creating order: " + err.Error())
	}

	procedure, err := e.Procedure.ToDB(ctx, site.ID, db)
	if err != nil {
		return errors.New("error creating procedure: " + err.Error())
	}

	exam, err := e.ToDB(ctx, order.ID, visit.ID, mrn.ID, site.ID, procedure.ID, order.CurrentStatus, db)
	if err != nil {
		return errors.New("error creating exam: " + err.Error())
	}

	reportModel := oru.getReport(exam, mrn, physician)
	if _, err = db.CreateReport(ctx, database.CreateReportParams{
		ExamID:        pgtype.Int8{Int64: int64(exam.ID), Valid: true},
		RadiologistID: pgtype.Int8{Int64: int64(physician.ID), Valid: true},
		Body:          reportModel.Body,
		Impression:    reportModel.Impression,
		ReportStatus:  reportModel.Status.String(),
		SubmittedDt:   pgtype.Timestamp{Time: reportModel.SubmittedDT, Valid: true},
	}); err != nil {
		return errors.New("error creating report: " + err.Error())
	}

	return nil
}

func (oru *ORU) getReport(exam database.Exam, mrn database.Mrn, radiologist database.Physician) entity.Report {
	body := ""

	for _, obx := range oru.OBX {
		if obx.ObservationValue != "" {
			body += obx.ObservationValue + "\n"
		}
	}

	observation := ""
	if len(oru.OBX) > 0 {
		observation = oru.OBX[0].ObservationValue
	}

	submitDT, err := time.Parse("20060102150405", oru.OBR.ObservationDT)
	if err != nil {
		submitDT = time.Now()
	}

	mrnModel := entity.MRN{
		Base: entity.Base{
			ID:        int(mrn.ID),
			CreatedAt: mrn.CreatedAt.Time,
			UpdatedAt: mrn.UpdatedAt.Time,
		},
		Value: mrn.Mrn,
	}

	examModel := entity.Exam{
		Base: entity.Base{
			ID:        int(exam.ID),
			CreatedAt: exam.CreatedAt.Time,
			UpdatedAt: exam.UpdatedAt.Time,
		},
		Accession: exam.Accession,
		MRN:       mrnModel,
	}

	radModel := entity.Physician{
		Base: entity.Base{
			ID:        int(radiologist.ID),
			CreatedAt: radiologist.CreatedAt.Time,
			UpdatedAt: radiologist.UpdatedAt.Time,
		},
		Name: objects.Name{
			Last:   radiologist.LastName,
			First:  radiologist.FirstName,
			Middle: radiologist.MiddleName.String,
			Suffix: radiologist.Suffix.String,
			Prefix: radiologist.Prefix.String,
			Degree: radiologist.Degree.String,
		},
		NPI:       radiologist.Npi,
		Specialty: objects.Specialty(radiologist.Specialty.String),
	}

	return entity.Report{
		Exam:        examModel,
		Radiologist: radModel,
		Body:        body,
		Impression:  observation,
		Status:      objects.NewReportStatus(oru.OBR.Status),
		SubmittedDT: submitDT,
	}
}
