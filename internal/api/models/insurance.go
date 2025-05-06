package models

type InsuranceModel struct {
	SetID          string `hl7:"IN1.1"`
	PlanID         string `hl7:"IN1.2"`
	CompanyName    string `hl7:"IN1.4"`
	CompanyAddress string `hl7:"IN1.5"`
	CompanyPhone   string `hl7:"IN1.7"`
	GroupNumber    string `hl7:"IN1.8"`
	PolicyNumber   string `hl7:"IN1.36"`
}
