package main

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type MessageModel struct {
	FieldSeparator string `json:"MSH.1"`
	EncodingChars  string `json:"MSH.2"`
	SendingApp     string `json:"MSH.3"`
	SendingFac     string `json:"MSH.4"`
	ReceivingApp   string `json:"MSH.5"`
	ReceivingFac   string `json:"MSH.6"`
	DateTime       string `json:"MSH.7"`
	Type           CM_MSG `json:"MSH.9"`
	ControlID      string `json:"MSH.10"`
	ProcessingID   string `json:"MSH.11"`
	Version        string `json:"MSH.12"`
}

func (m *MessageModel) ToEntity() entity.Message {
	dt, err := time.Parse("20060102150405", m.DateTime)
	if err != nil {
		dt = time.Now()
	}
	return entity.Message{
		FieldSeparator: m.FieldSeparator,
		EncodingChars:  m.EncodingChars,
		SendingApp:     m.SendingApp,
		SendingFac:     m.SendingFac,
		ReceivingApp:   m.ReceivingApp,
		ReceivingFac:   m.ReceivingFac,
		DateTime:       dt,
		Type:           m.Type.Type,
		TriggerEvent:   m.Type.TriggerEvent,
		ControlID:      m.ControlID,
		ProcessingID:   m.ProcessingID,
		Version:        m.Version,
	}
}

type PatientModel struct {
	MRN  CX     `json:"PID.3"`
	Name XPN    `json:"PID.5"`
	DOB  string `json:"PID.7"`
	Sex  string `json:"PID.8"`
	SSN  string `json:"PID.19"`
}

func (p *PatientModel) ToEntity() entity.Patient {
	name := objects.Name{
		Last:   p.Name.LastName,
		First:  p.Name.FirstName,
		Middle: p.Name.MiddleName,
		Suffix: p.Name.Suffix,
		Prefix: p.Name.Prefix,
		Degree: p.Name.Degree,
	}

	return entity.Patient{
		Name: name,
		DOB:  tryParseDOB(p.DOB),
		Sex:  p.Sex,
		SSN:  p.SSN,
	}

}

func tryParseDOB(dob string) time.Time {
	// try to parse dob a few different ways
	// if none work, use current time
	formats := []string{
		"20060102",
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"01-02-2006",
		"01/02/06",
	}

	for _, f := range formats {
		dt, err := time.Parse(f, dob)
		if err == nil {
			return dt
		}
	}

	return time.Now()
}

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

type OrderModel struct {
	OrderNo          string `json:"ORC.2"`
	FillerOrderNo    string `json:"ORC.3"`
	OrderDT          string `json:"ORC.9"`
	OrderingProvider XCN    `json:"ORC.12"`
}

func (o *OrderModel) ToEntity() entity.Order {
	orderDT, err := time.Parse("20060102150405", o.OrderDT)
	if err != nil {
		orderDT = time.Now()
	}

	provider := entity.Physician{
		Name: objects.Name{
			Last:   o.OrderingProvider.FamilyName,
			First:  o.OrderingProvider.GivenName,
			Middle: o.OrderingProvider.MiddleName,
			Suffix: o.OrderingProvider.Suffix,
			Prefix: o.OrderingProvider.Prefix,
			Degree: o.OrderingProvider.Degree,
		},
		// TODO: NPI
	}

	return entity.Order{
		Accession: o.FillerOrderNo,
		Date:      orderDT,
		Provider:  provider,
	}
}

type ExamModel struct {
	Accession     string `json:"OBR.3"`
	Service       CE     `json:"OBR.4"`
	Priority      string `json:"OBR.5"`
	RequestDT     string `json:"OBR.6"`
	ObservationDT string `json:"OBR.7"`
	StatusDT      string `json:"OBR.22"`
}

func (e *ExamModel) ToEntity(siteCode string, status string, mrn CX) entity.Exam {
	procedure := entity.Procedure{
		Code:        e.Service.Identifier,
		Description: e.Service.Text,
	}

	site := entity.Site{Code: siteCode}

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

type InsuranceModel struct {
	SetID          string `json:"IN1.1"`
	PlanID         string `json:"IN1.2"`
	CompanyName    string `json:"IN1.4"`
	CompanyAddress string `json:"IN1.5"`
	CompanyPhone   string `json:"IN1.7"`
	GroupNumber    string `json:"IN1.8"`
	PolicyNumber   string `json:"IN1.36"`
}

type EventModel struct {
	Code       string `json:"EVN.1"`
	DT         string `json:"EVN.2"`
	OperatorID XCN    `json:"EVN.5"`
	OccurredDT string `json:"EVN.6"`
}

type ReportModel struct {
	SetID            string `json:"OBX.1"`
	ValueType        string `json:"OBX.2"`
	Service          CE     `json:"OBX.3"`
	ObservationSubID string `json:"OBX.4"`
	ObservationValue string `json:"OBX.5"`
	ResultStatus     string `json:"OBX.11"`
	ObservationDT    string `json:"OBX.14"`
}
