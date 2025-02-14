package models

type InsuranceModel struct {
	SetID          string `json:"IN1.1"`
	PlanID         string `json:"IN1.2"`
	CompanyName    string `json:"IN1.4"`
	CompanyAddress string `json:"IN1.5"`
	CompanyPhone   string `json:"IN1.7"`
	GroupNumber    string `json:"IN1.8"`
	PolicyNumber   string `json:"IN1.36"`
}
