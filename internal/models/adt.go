package models

import "encoding/json"

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
