package models

import "encoding/json"

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
