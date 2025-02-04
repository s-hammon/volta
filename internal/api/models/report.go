package models

type ReportModel struct {
	SetID            string `hl7:"OBX.1"`
	ValueType        string `hl7:"OBX.2"`
	Service          CE     `hl7:"OBX.3"`
	ObservationSubID string `hl7:"OBX.4"`
	ObservationValue string `hl7:"OBX.5"`
	ResultStatus     string `hl7:"OBX.11"`
	ObservationDT    string `hl7:"OBX.14"`
}
