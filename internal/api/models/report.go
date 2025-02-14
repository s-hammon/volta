package models

type ReportModel struct {
	SetID            string `json:"OBX.1"`
	ValueType        string `json:"OBX.2"`
	Service          CE     `json:"OBX.3"`
	ObservationSubID string `json:"OBX.4"`
	ObservationValue string `json:"OBX.5"`
	ResultStatus     string `json:"OBX.11"`
	ObservationDT    string `json:"OBX.14"`
}
