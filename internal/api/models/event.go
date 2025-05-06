package models

type EventModel struct {
	Code       string `hl7:"EVN.1"`
	DT         string `hl7:"EVN.2"`
	OperatorID XCN    `hl7:"EVN.5"`
	OccurredDT string `hl7:"EVN.6"`
}
