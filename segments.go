package main

import (
	"encoding/json"
)

type ADT struct {
	MSH MessageModel   `json:"MSH"`
	EVN EventModel     `json:"EVN"`
	PID PatientModel   `json:"PID"`
	PV1 VisitModel     `json:"PV1"`
	IN1 InsuranceModel `json:"IN1"`
}

func NewADT(msgMap map[string]interface{}) (ADT, error) {
	b, err := json.Marshal(msgMap)
	if err != nil {
		return ADT{}, err
	}

	adt := ADT{}
	if err = json.Unmarshal(b, &adt); err != nil {
		return ADT{}, err
	}

	return adt, nil
}

type ORM struct {
	MSH MessageModel `json:"MSH"`
	PID PatientModel `json:"PID"`
	PV1 VisitModel   `json:"PV1"`
	ORC OrderModel   `json:"ORC"`
	OBR ExamModel    `json:"OBR"`
}

func NewORM(msgMap map[string]interface{}) (ORM, error) {
	b, err := json.Marshal(msgMap)
	if err != nil {
		return ORM{}, err
	}

	orm := ORM{}
	if err = json.Unmarshal(b, &orm); err != nil {
		return ORM{}, err
	}

	return orm, nil
}

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
