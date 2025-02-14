package models

// Messate Type
type CM_MSG struct {
	Type         string `json:"1"`
	TriggerEvent string `json:"2"`
}

// Extended Person Name
type XPN struct {
	LastName   string `json:"1"`
	FirstName  string `json:"2"`
	MiddleName string `json:"3"`
	Suffix     string `json:"4"`
	Prefix     string `json:"5"`
	Degree     string `json:"6"`
}

// Extended Composite ID
type CX struct {
	ID                 string `json:"1"`
	CheckDigit         string `json:"2"`
	CheckDigitScheme   string `json:"3"`
	AssigningAuthority string `json:"4"`
	IdentifierTypeCode string `json:"5"`
}

// Person Location
type PL struct {
	PointOfCare         string `json:"1"`
	Room                string `json:"2"`
	Bed                 string `json:"3"`
	Facility            string `json:"4"`
	LocationStatus      string `json:"5"`
	PersonLocationType  string `json:"6"`
	Building            string `json:"7"`
	Floor               string `json:"8"`
	LocationDescription string `json:"9"`
}

type EI struct {
	EntityIdentifier string `json:"1"`
	NamespaceID      string `json:"2"`
	UniversalID      string `json:"3"`
	UniversalIDType  string `json:"4"`
}

// Hierarchic Designator
type HD struct {
	NamespaceID     string `json:"1"`
	UniversalID     string `json:"2"`
	UniversalIDType string `json:"3"`
}

// Extended Composite ID & Name
type XCN struct {
	IDNumber   string `json:"1"`
	FamilyName string `json:"2"`
	GivenName  string `json:"3"`
	MiddleName string `json:"4"`
	Suffix     string `json:"5"`
	Prefix     string `json:"6"`
	Degree     string `json:"7"`
}

// Coded Element
type CE struct {
	Identifier      string `json:"1"`
	Text            string `json:"2"`
	CodingSystem    string `json:"3"`
	AltIdentifier   string `json:"4"`
	AltText         string `json:"5"`
	AltCodingSystem string `json:"6"`
}

// Extended Address
type XAD struct {
	StreetAddress    string `json:"1"`
	OtherDesignation string `json:"2"`
	City             string `json:"3"`
	State            string `json:"4"`
	Zip              string `json:"5"`
	Country          string `json:"6"`
}
