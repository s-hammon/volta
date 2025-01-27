package models

type EventModel struct {
	Code       string `json:"EVN.1"`
	DT         string `json:"EVN.2"`
	OperatorID XCN    `json:"EVN.5"`
	OccurredDT string `json:"EVN.6"`
}
