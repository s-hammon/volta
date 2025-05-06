package models

// Messate Type
type CM_MSG struct {
	Name         string `hl7:"1"`
	TriggerEvent string `hl7:"2"`
}

// Extended Person Name
type XPN struct {
	LastName   string `hl7:"1"`
	FirstName  string `hl7:"2"`
	MiddleName string `hl7:"3"`
	Suffix     string `hl7:"4"`
	Prefix     string `hl7:"5"`
	Degree     string `hl7:"6"`
}

// Extended Composite ID
type CX struct {
	ID                 string `hl7:"1"`
	CheckDigit         string `hl7:"2"`
	CheckDigitScheme   string `hl7:"3"`
	AssigningAuthority string `hl7:"4"`
	IdentifierTypeCode string `hl7:"5"`
}

// Person Location
type PL struct {
	PointOfCare         string `hl7:"1"`
	Room                string `hl7:"2"`
	Bed                 string `hl7:"3"`
	Facility            string `hl7:"4"`
	LocationStatus      string `hl7:"5"`
	PersonLocationType  string `hl7:"6"`
	Building            string `hl7:"7"`
	Floor               string `hl7:"8"`
	LocationDescription string `hl7:"9"`
}

// Entity Identifier
type EI struct {
	EntityIdentifier string `hl7:"1"`
	NamespaceID      string `hl7:"2"`
	UniversalID      string `hl7:"3"`
	UniversalIDType  string `hl7:"4"`
}

// Hierarchic Designator
type HD struct {
	NamespaceID     string `hl7:"1"`
	UniversalID     string `hl7:"2"`
	UniversalIDType string `hl7:"3"`
}

// Extended Composite ID & Name
type XCN struct {
	IDNumber   string `hl7:"1"`
	FamilyName string `hl7:"2"`
	GivenName  string `hl7:"3"`
	MiddleName string `hl7:"4"`
	Suffix     string `hl7:"5"`
	Prefix     string `hl7:"6"`
	Degree     string `hl7:"7"`
}

// Coded Element
type CE struct {
	Identifier      string `hl7:"1"`
	Text            string `hl7:"2"`
	CodingSystem    string `hl7:"3"`
	AltIdentifier   string `hl7:"4"`
	AltText         string `hl7:"5"`
	AltCodingSystem string `hl7:"6"`
}

// Extended Address
type XAD struct {
	StreetAddress    string `hl7:"1"`
	OtherDesignation string `hl7:"2"`
	City             string `hl7:"3"`
	State            string `hl7:"4"`
	Zip              string `hl7:"5"`
	Country          string `hl7:"6"`
}

// Observing Practitioner (i.e. radiologist)
type CM_NDL struct {
	ObservingPractitioner XCN    `hl7:"1"`
	ObservationDT         string `hl7:"3"`
}
