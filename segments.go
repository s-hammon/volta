package main

import (
	"encoding/json"
	"fmt"
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

func (a *ADT) Materialize() {
	// check if MSH is not empty
	if a.MSH != (MessageModel{}) {
		msg := a.MSH.ToEntity()
		fmt.Println(msg)
	}
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

func (o *ORM) Materialize() {
	siteCode := o.MSH.SendingFac
	mrn := o.PID.MRN
	status := o.PV1.Class

	msg := o.MSH.ToEntity()
	pt := o.PID.ToEntity()
	visit := o.PV1.ToEntity(siteCode, mrn)
	order := o.ORC.ToEntity()
	exam := o.OBR.ToEntity(siteCode, status, mrn)

	// print all details
	fmt.Printf("Message:\t%+v\n", msg)
	fmt.Printf("Patient:\t%+v\n", pt)
	fmt.Printf("Visit:\t%+v\n", visit)
	fmt.Printf("Order:\t%+v\n", order)
	fmt.Printf("Exam:\t%+v\n", exam)
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
