package api

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

const cstName = "America/Chicago"

var cst, _ = time.LoadLocation(cstName)

var dateFormats = []string{
	"20060102",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"01-02-2006",
	"01/02/06",
}

type Message struct {
	FieldSeparator string `hl7:"MSH.1"`
	EncodingChars  string `hl7:"MSH.2"`
	SendingApp     string `hl7:"MSH.3"`
	SendingFac     string `hl7:"MSH.4"`
	ReceivingApp   string `hl7:"MSH.5"`
	ReceivingFac   string `hl7:"MSH.6"`
	DateTime       string `hl7:"MSH.7"`
	MsgType        CM_MSG `hl7:"MSH.9"`
	ControlID      string `hl7:"MSH.10"`
	ProcessingID   string `hl7:"MSH.11"`
	Version        string `hl7:"MSH.12"`
}

type ORM struct {
	FieldSeparator   string `hl7:"MSH.1"`
	EncodingChars    string `hl7:"MSH.2"`
	SendingApp       string `hl7:"MSH.3"`
	SendingFac       string `hl7:"MSH.4"`
	ReceivingApp     string `hl7:"MSH.5"`
	ReceivingFac     string `hl7:"MSH.6"`
	DateTime         string `hl7:"MSH.7"`
	MsgType          CM_MSG `hl7:"MSH.9"`
	ControlID        string `hl7:"MSH.10"`
	ProcessingID     string `hl7:"MSH.11"`
	Version          string `hl7:"MSH.12"`
	MRN              CX     `hl7:"PID.3"`
	PatientName      XPN    `hl7:"PID.5"`
	DOB              string `hl7:"PID.7"`
	Sex              string `hl7:"PID.8"`
	SSN              string `hl7:"PID.19"`
	VisitNo          string `hl7:"PV1.19"`
	PatientClass     string `hl7:"PV1.2"`
	AssignedLocation PL     `hl7:"PV1.3"`
	Accession        string `hl7:"ORC.2"`
	FillerOrderNo    string `hl7:"ORC.3"`
	OrderStatus      string `hl7:"ORC.5"`
	OrderDT          string `hl7:"ORC.9"`
	OrderingProvider XCN    `hl7:"ORC.12"`
	Service          CE     `hl7:"OBR.4"`
	StatusDT         string `hl7:"OBR.22"`
}

func (o *ORM) ToOrder() *entity.Order {
	site := &entity.Site{Code: o.SendingFac}
	order := &entity.Order{}
	order.Message = entity.Message{
		FieldSeparator: o.FieldSeparator,
		EncodingChars:  o.EncodingChars,
		SendingApp:     o.SendingApp,
		SendingFac:     o.SendingFac,
		ReceivingApp:   o.ReceivingApp,
		ReceivingFac:   o.ReceivingFac,
		DateTime:       convertCSTtoUTC(o.DateTime),
		Type:           o.MsgType.Name,
		TriggerEvent:   o.MsgType.TriggerEvent,
		ControlID:      o.ControlID,
		ProcessingID:   o.ProcessingID,
		Version:        o.Version,
	}
	order.Patient = entity.Patient{
		Name: objects.Name{
			Last:   o.PatientName.LastName,
			First:  o.PatientName.FirstName,
			Middle: o.PatientName.MiddleName,
			Suffix: o.PatientName.Suffix,
			Prefix: o.PatientName.Prefix,
			Degree: o.PatientName.Degree,
		},
		DOB: tryParseDOB(o.DOB),
		Sex: o.Sex,
		SSN: objects.NewSSN(o.SSN),
	}
	order.Visit = entity.Visit{
		VisitNo: o.VisitNo,
		Site:    *site,
		MRN: entity.MRN{
			Value:              o.MRN.ID,
			AssigningAuthority: o.MRN.AssigningAuthority,
		},
		Type: objects.NewPatientType(o.PatientClass),
	}
	order.Procedure = entity.Procedure{
		Site:        *site,
		Code:        o.Service.Identifier,
		Description: o.Service.Text,
	}
	order.Exam = entity.Exam{
		Accession: coalesce(o.Accession, o.FillerOrderNo),
		Procedure: entity.Procedure{
			Site:        *site,
			Code:        o.Service.Identifier,
			Description: o.Service.Text,
		},
		CurrentStatus: entity.NewExamStatus(o.OrderStatus),
		Provider: entity.Physician{
			Name: objects.Name{
				Last:   o.OrderingProvider.LastName,
				First:  o.OrderingProvider.FirstName,
				Middle: o.OrderingProvider.MiddleName,
				Suffix: o.OrderingProvider.Suffix,
				Prefix: o.OrderingProvider.Prefix,
				Degree: o.OrderingProvider.Degree,
			},
		},
		Site: *site,
	}
	dt := convertCSTtoUTC(o.OrderDT)
	switch order.Exam.CurrentStatus {
	case "SC":
		order.Exam.Scheduled = dt
	case "IP":
		order.Exam.Begin = dt
	case "CM":
		order.Exam.End = dt
	case "CA":
		order.Exam.Cancelled = dt
	}
	return order
}

type ORU struct {
	FieldSeparator   string `hl7:"MSH.1"`
	EncodingChars    string `hl7:"MSH.2"`
	SendingApp       string `hl7:"MSH.3"`
	SendingFac       string `hl7:"MSH.4"`
	ReceivingApp     string `hl7:"MSH.5"`
	ReceivingFac     string `hl7:"MSH.6"`
	DateTime         string `hl7:"MSH.7"`
	MsgType          CM_MSG `hl7:"MSH.9"`
	ControlID        string `hl7:"MSH.10"`
	ProcessingID     string `hl7:"MSH.11"`
	Version          string `hl7:"MSH.12"`
	MRN              CX     `hl7:"PID.3"`
	PatientName      XPN    `hl7:"PID.5"`
	DOB              string `hl7:"PID.7"`
	Sex              string `hl7:"PID.8"`
	SSN              string `hl7:"PID.19"`
	VisitNo          string `hl7:"PV1.19"`
	PatientClass     string `hl7:"PV1.2"`
	AssignedLocation PL     `hl7:"PV1.3"`
}

func (o *ORU) ToObservation(report entity.Report, exams ...Exam) *entity.Observation {
	observation := &entity.Observation{}
	site := &entity.Site{Code: o.SendingFac}
	observation.Message = entity.Message{
		FieldSeparator: o.FieldSeparator,
		EncodingChars:  o.EncodingChars,
		SendingApp:     o.SendingApp,
		SendingFac:     o.SendingFac,
		ReceivingApp:   o.ReceivingApp,
		ReceivingFac:   o.ReceivingFac,
		DateTime:       convertCSTtoUTC(o.DateTime),
		Type:           o.MsgType.Name,
		TriggerEvent:   o.MsgType.TriggerEvent,
		ControlID:      o.ControlID,
		ProcessingID:   o.ProcessingID,
		Version:        o.Version,
	}
	observation.Patient = entity.Patient{
		Name: objects.Name{
			Last:   o.PatientName.LastName,
			First:  o.PatientName.FirstName,
			Middle: o.PatientName.MiddleName,
			Suffix: o.PatientName.Suffix,
			Prefix: o.PatientName.Prefix,
			Degree: o.PatientName.Degree,
		},
		DOB: tryParseDOB(o.DOB),
		Sex: o.Sex,
		SSN: objects.NewSSN(o.SSN),
	}
	observation.Visit = entity.Visit{
		VisitNo: o.VisitNo,
		Site:    *site,
		MRN: entity.MRN{
			Value:              o.MRN.ID,
			AssigningAuthority: o.MRN.AssigningAuthority,
		},
	}
	observation.Report = report
	for _, exam := range exams {
		observation.Exams = append(observation.Exams, exam.ToEntity(*site))
	}
	return observation
}

type Exam struct {
	Accession        string `hl7:"ORC.2"`
	FillerOrderNo    string `hl7:"ORC.3"`
	OrderStatus      string `hl7:"ORC.5"`
	OrderDT          string `hl7:"ORC.9"`
	OrderingProvider XCN    `hl7:"ORC.12"`
	Service          CE     `hl7:"OBR.4"`
	StatusDT         string `hl7:"OBR.22"`
}

func (e *Exam) ToEntity(site entity.Site) (exam entity.Exam) {
	exam.Accession = coalesce(e.Accession, e.FillerOrderNo)
	exam.Procedure = entity.Procedure{
		Site:        site,
		Code:        e.Service.Identifier,
		Description: e.Service.Text,
	}
	exam.CurrentStatus = entity.NewExamStatus(e.OrderStatus)
	exam.Provider = entity.Physician{
		Name: objects.Name{
			Last:   e.OrderingProvider.LastName,
			First:  e.OrderingProvider.FirstName,
			Middle: e.OrderingProvider.MiddleName,
			Suffix: e.OrderingProvider.Suffix,
			Prefix: e.OrderingProvider.Prefix,
			Degree: e.OrderingProvider.Degree,
		},
	}
	exam.Site = site
	return exam
}

type Report struct {
	Radiologist      CM_NDL `hl7:"OBR.32"`
	SetID            string `hl7:"OBX.1"`
	ValueType        string `hl7:"OBX.2"`
	Service          CE     `hl7:"OBX.3"`
	ObservationSubID string `hl7:"OBX.4"`
	ObservationValue string `hl7:"OBX.5"`
	ResultStatus     string `hl7:"OBX.11"`
	ObservationDT    string `hl7:"OBX.14"`
}

func GetReport(obx []Report) (r entity.Report) {
	if len(obx) > 0 {
		rad := obx[0].Radiologist.ObservingPractitioner
		r.Radiologist = entity.Physician{
			Name: objects.NewName(
				rad.LastName,
				rad.FirstName,
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
	for _, o := range obx {
		if o.ObservationValue != "" {
			r.Body += o.ObservationValue + "\n"
		}
	}
	return r
}

// Messate Type
type CM_MSG struct {
	Name         string `hl7:"1"`
	TriggerEvent string `hl7:"2"`
}

// Extended Composite ID
type CX struct {
	ID                 string `hl7:"1"`
	CheckDigit         string `hl7:"2"`
	CheckDigitScheme   string `hl7:"3"`
	AssigningAuthority string `hl7:"4"`
	IdentifierTypeCode string `hl7:"5"`
}

// Extended Person Name
type XPN struct {
	LastName   string `hl7:"1"`
	FirstName  string `hl7:"2"`
	MiddleName string `hl7:"3"`
	Suffix     string `hl7:"4"`
	Prefix     string `hl7:"5"`
	Degree     string `hl7:"6"`
}

// Person Location
type PL struct {
	PointOfCare         string `hl7:"1"`
	Room                string `hl7:"2"`
	Bed                 string `hl7:"3"`
	Facility            string `hl7:"4"`
	LocationStatus      string `hl7:"5"`
	PersonLocationType  string `hl7:"6"`
	Building            string `hl7:"7"`
	Floor               string `hl7:"8"`
	LocationDescription string `hl7:"9"`
}

// Extended Composite ID & Name
type XCN struct {
	IDNumber   string `hl7:"1"`
	LastName   string `hl7:"2"`
	FirstName  string `hl7:"3"`
	MiddleName string `hl7:"4"`
	Suffix     string `hl7:"5"`
	Prefix     string `hl7:"6"`
	Degree     string `hl7:"7"`
}

// Coded Element
type CE struct {
	Identifier      string `hl7:"1"`
	Text            string `hl7:"2"`
	CodingSystem    string `hl7:"3"`
	AltIdentifier   string `hl7:"4"`
	AltText         string `hl7:"5"`
	AltCodingSystem string `hl7:"6"`
}

// Observing Practitioner (i.e. radiologist)
type CM_NDL struct {
	ObservingPractitioner XCN    `hl7:"1"`
	ObservationDT         string `hl7:"3"`
}

func tryParseDOB(dob string) time.Time {
	// try to parse dob a few different ways
	// if none work, use current time
	for _, f := range dateFormats {
		dt, err := time.Parse(f, dob)
		if err == nil {
			return dt
		}
	}

	return time.Now()
}

func convertCSTtoUTC(stringDT string) time.Time {
	dt, err := time.ParseInLocation("20060102150405", stringDT, cst)
	if err != nil {
		dt, err = time.Parse("20060102150405", stringDT)
		if err != nil {
			dt = time.Now()
		}
	}

	return dt.UTC()
}

func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
