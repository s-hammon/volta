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
	MSH MessageModel  `json:"MSH"`
	PID PatientModel  `json:"PID"`
	PV1 VisitModel    `json:"PV1"`
	ORC []OrderModel  `json:"ORC"`
	OBR []ExamModel   `json:"OBR"`
	OBX []ReportModel `json:"OBX"`
}

func (oru *ORU) ToDB(ctx context.Context, db *database.Queries) error {
	p := oru.PID.ToEntity()
	m := oru.MSH.ToEntity()
	v := oru.PV1.ToEntity(m.SendingFac, oru.PID.MRN)

	orderGroups, err := oru.groupOrders()
	if err != nil {
		return errors.New("error grouping orders and exams: " + err.Error())
	}
	oe := newOrderEntities(v.Site.Code, oru.PID.MRN, orderGroups...)

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

	physician, err := oe[0].order.Provider.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating physician: " + err.Error())
	}

	examIDs := make([]int64, len(oe))
	for i, orderEntity := range oe {
		order, err := orderEntity.order.ToDB(ctx, site.ID, visit.ID, mrn.ID, physician.ID, db)
		if err != nil {
			return errors.New("error creating order: " + err.Error())
		}

		procedure, err := orderEntity.exam.Procedure.ToDB(ctx, site.ID, db)
		if err != nil {
			return errors.New("error creating procedure: " + err.Error())
		}

		exam, err := orderEntity.exam.ToDB(ctx, order.ID, visit.ID, mrn.ID, site.ID, procedure.ID, order.CurrentStatus, db)
		if err != nil {
			return errors.New("error creating exam: " + err.Error())
		}
		examIDs[i] = exam.ID
	}

	reportModel := oru.getReport(physician)
	report, err := db.CreateReport(ctx, database.CreateReportParams{
		RadiologistID: pgtype.Int8{Int64: int64(physician.ID), Valid: true},
		Body:          reportModel.Body,
		Impression:    reportModel.Impression,
		ReportStatus:  reportModel.Status.String(),
		SubmittedDt:   pgtype.Timestamp{Time: reportModel.SubmittedDT, Valid: true},
	})
	if err != nil {
		return errors.New("error creating report: " + err.Error())
	}

	switch reportModel.Status {
	case objects.Final:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamFinalReport(ctx, database.UpdateExamFinalReportParams{
				ID:            int64(examID),
				FinalReportID: pgtype.Int8{Int64: report.ID, Valid: true},
			}); err != nil {
				return errors.New("error updating exam with final report: " + err.Error())
			}
		}
	case objects.Addendum:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamAddendumReport(ctx, database.UpdateExamAddendumReportParams{
				ID:               int64(examID),
				AddendumReportID: pgtype.Int8{Int64: examID, Valid: true},
			}); err != nil {
				return errors.New("error updating exam with addendum report: " + err.Error())
			}
		}
	}

	return nil
}

func (oru *ORU) getReport(radiologist database.Physician) entity.Report {
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

	submitDT, err := time.Parse("20060102150405", oru.OBR[0].ObservationDT) // TODO: audit this, seems like times are all over the place in HL7
	if err != nil {
		submitDT = time.Now()
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
		Radiologist: radModel,
		Body:        body,
		Impression:  observation,
		Status:      objects.NewReportStatus(oru.OBR[0].Status), // TODO: also audit this
		SubmittedDT: submitDT,
	}
}

type orderGroup struct {
	Order OrderModel
	Exam  ExamModel
}

func (oru *ORU) groupOrders() ([]orderGroup, error) {
	if len(oru.ORC) == 0 || len(oru.OBR) == 0 {
		return nil, errors.New("no orders or exams to group")
	}
	if len(oru.ORC) != len(oru.OBR) {
		return nil, errors.New("mismatched number of orders and exams")
	}

	groups := make([]orderGroup, len(oru.ORC))
	for i, o := range oru.ORC {
		groups[i] = orderGroup{Order: o}
		// TODO: vet this logic
		acc := coalesce(o.OrderNo, o.FillerOrderNo)
		for _, e := range oru.OBR {
			if acc == e.Accession {
				groups[i].Exam = e
				break
			}
		}
		// make sure we indeed found an exam for this order
		if groups[i].Exam == (ExamModel{}) {
			return nil, errors.New("no exam found for order number " + acc)
		}
	}

	return groups, nil
}

type orderEntity struct {
	order entity.Order
	exam  entity.Exam
}

func newOrderEntities(visitSiteCode string, mrn CX, orderGroups ...orderGroup) []orderEntity {
	entities := make([]orderEntity, len(orderGroups))
	for i, group := range orderGroups {
		o := group.Order.ToEntity()
		e := group.Exam.ToEntity(visitSiteCode, o.CurrentStatus, mrn)
		entities[i] = orderEntity{order: o, exam: e}
	}

	return entities
}

// if a is empty, return b
func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
