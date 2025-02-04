package models

type ADT struct {
	MSH MessageModel   `hl7:"MSH"`
	EVN EventModel     `hl7:"EVN"`
	PID PatientModel   `hl7:"PID"`
	PV1 VisitModel     `hl7:"PV1"`
	IN1 InsuranceModel `hl7:"IN1"`
}
