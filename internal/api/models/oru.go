package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

func (oru *ORU) UnmarshalJSON(data []byte) error {
	type Alias ORU
	aux := &struct {
		ORC json.RawMessage `json:"ORC"`
		OBR json.RawMessage `json:"OBR"`
		OBX json.RawMessage `json:"OBX"`
		*Alias
	}{
		Alias: (*Alias)(oru),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ORC != nil {
		orc, err := normalizeToSlice[OrderModel](aux.ORC)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ORC: %w", err)
		}
		oru.ORC = orc
	}
	if aux.OBR != nil {
		obr, err := normalizeToSlice[ExamModel](aux.OBR)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ORC: %w", err)
		}
		oru.OBR = obr
	}
	if aux.OBX != nil {
		obx, err := normalizeToSlice[ReportModel](aux.OBX)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ORC: %w", err)
		}
		oru.OBX = obx
	}

	return nil
}

func (oru *ORU) ToDB(ctx context.Context, db *database.Queries) error {
	p := oru.PID.ToEntity()
	m := oru.MSH.ToEntity()
	v := oru.PV1.ToEntity(m.SendingFac, oru.PID.MRN)

	orderGroups, err := oru.GroupOrders()
	if err != nil {
		return errors.New("error grouping orders and exams: " + err.Error())
	}
	oe := NewOrderEntities(v.Site.Code, oru.PID.MRN, orderGroups...)

	siteID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating site: " + err.Error())
	}

	patientID, err := p.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating patient: " + err.Error())
	}

	mrnID, err := v.MRN.ToDB(ctx, siteID, patientID, db)
	if err != nil {
		return errors.New("error creating mrn: " + err.Error())
	}

	visitID, err := v.ToDB(ctx, siteID, mrnID, db)
	if err != nil {
		return errors.New("error creating visit: " + err.Error())
	}

	physicianID, err := oe[0].order.Provider.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating physician: " + err.Error())
	}

	examIDs := make([]int64, len(oe))
	for i, orderEntity := range oe {
		orderEntity.order.CurrentStatus = entity.OrderComplete
		orderID, orderStatus, err := orderEntity.order.ToDB(ctx, siteID, visitID, mrnID, physicianID, db)
		if err != nil {
			return errors.New("error creating order: " + err.Error())
		}

		procedureID, err := orderEntity.exam.Procedure.ToDB(ctx, siteID, db)
		if err != nil {
			return errors.New("error creating procedure: " + err.Error())
		}

		examID, err := orderEntity.exam.ToDB(ctx, orderID, visitID, mrnID, siteID, procedureID, orderStatus, db)
		if err != nil {
			return errors.New("error creating exam: " + err.Error())
		}
		examIDs[i] = examID
	}

	reportModel := oru.GetReport()
	report, err := db.CreateReport(ctx, database.CreateReportParams{
		RadiologistID: pgtype.Int8{Int64: int64(physicianID), Valid: true},
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
				AddendumReportID: pgtype.Int8{Int64: report.ID, Valid: true},
			}); err != nil {
				return errors.New("error updating exam with addendum report: " + err.Error())
			}
		}
	case objects.Pending:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamPrelimReport(ctx, database.UpdateExamPrelimReportParams{
				ID:             int64(examID),
				PrelimReportID: pgtype.Int8{Int64: report.ID, Valid: true},
			}); err != nil {
				return errors.New("error updating exam with correction report: " + err.Error())
			}
		}
	}

	return nil
}

func (oru *ORU) GetReport() entity.Report {
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
	submitDT := convertCSTtoUTC(oru.OBR[0].StatusDT)
	return entity.Report{
		Body:        body,
		Impression:  observation,
		Status:      objects.NewReportStatus(oru.OBR[0].Status),
		SubmittedDT: submitDT,
	}
}

type orderGroup struct {
	Order OrderModel
	Exam  ExamModel
}

func (oru *ORU) GroupOrders() ([]orderGroup, error) {
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

func NewOrderEntities(visitSiteCode string, mrn CX, orderGroups ...orderGroup) []orderEntity {
	entities := make([]orderEntity, len(orderGroups))
	for i, group := range orderGroups {
		o := group.Order.ToEntity()
		e := group.Exam.ToEntity(visitSiteCode, o.CurrentStatus.String(), mrn)
		entities[i] = orderEntity{order: o, exam: e}
	}

	return entities
}

func (oe *orderEntity) GetOrder() entity.Order {
	return oe.order
}

func (oe *orderEntity) GetExam() entity.Exam {
	return oe.exam
}

// if a is empty, return b
func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func normalizeToSlice[T any](raw json.RawMessage) ([]T, error) {
	var slice []T
	if err := json.Unmarshal(raw, &slice); err == nil {
		return slice, nil
	}

	var single T
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, err
	}

	return []T{single}, nil
}
