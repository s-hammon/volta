package models

type ADT struct {
	MSH MessageModel   `json:"MSH"`
	EVN EventModel     `json:"EVN"`
	PID PatientModel   `json:"PID"`
	PV1 VisitModel     `json:"PV1"`
	IN1 InsuranceModel `json:"IN1"`
}
