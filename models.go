package main

import (
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

type PatientModel struct {
	ID   CX     `json:"PID.3"`
	Name []XPN  `json:"PID.5"`
	DOB  string `json:"PID.7"`
	Sex  string `json:"PID.8"`
	SSN  string `json:"PID.19"`
}

type VisitModel struct {
	VisitNo          string `json:"PV1.19"`
	AssignedLocation PL     `json:"PV1.3"`
}

func (m *PatientModel) GetMRN() entity.MRN {
	return entity.MRN{
		Value:              m.ID.ID,
		AssigningAuthority: m.ID.AssigningAuthority,
	}
}

func (m *PatientModel) ToPatientRepository() entity.Patient {
	// just get the first patient
	pt := m.Name[0]
	return entity.Patient{
		Name: objects.Name{
			Last:   pt.LastName,
			First:  pt.FirstName,
			Middle: pt.MiddleName,
		},
		DOB: m.DOB,
		SSN: m.SSN,
	}
}

type OrderModel struct {
	OrderNo          EI     `json:"ORC.2"`
	FillerOrderNo    EI     `json:"ORC.3"`
	OrderDT          string `json:"ORC.9"`
	OrderingProvider PL     `json:"ORC.12"`
}

type ExamModel struct {
	Accession     EI     `json:"OBR.3"`
	Service       CE     `json:"OBR.4"`
	Priority      string `json:"OBR.5"`
	RequestDT     string `json:"OBR.6"`
	ObservationDT string `json:"OBR.7"`
	StatusDT      string `json:"OBR.22"`
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
