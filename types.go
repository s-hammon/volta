package main

import (
	"encoding/json"
	"regexp"
)

var re = regexp.MustCompile(`\.(\d+)$`)

// Messate Type
type CM_MSG struct {
	Type         string `json:"1"`
	TriggerEvent string `json:"2"`
}

func (c *CM_MSG) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				c.Type = v
			case "2":
				c.TriggerEvent = v
			}
		}
	}

	return nil
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

func (e *XPN) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				e.LastName = v
			case "2":
				e.FirstName = v
			case "3":
				e.MiddleName = v
			case "4":
				e.Suffix = v
			case "5":
				e.Prefix = v
			case "6":
				e.Degree = v
			}
		}
	}

	return nil
}

// Extended Composite ID
type CX struct {
	ID                 string `json:"1"`
	CheckDigit         string `json:"2"`
	CheckDigitScheme   string `json:"3"`
	AssigningAuthority string `json:"4"`
	IdentifierTypeCode string `json:"5"`
}

func (i *CX) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				i.ID = v
			case "2":
				i.CheckDigit = v
			case "3":
				i.CheckDigitScheme = v
			case "4":
				i.AssigningAuthority = v
			case "5":
				i.IdentifierTypeCode = v
			}
		}
	}

	return nil
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

func (p *PL) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				p.PointOfCare = v
			case "2":
				p.Room = v
			case "3":
				p.Bed = v
			case "4":
				p.Facility = v
			case "5":
				p.LocationStatus = v
			case "6":
				p.PersonLocationType = v
			case "7":
				p.Building = v
			case "8":
				p.Floor = v
			case "9":
				p.LocationDescription = v
			}
		}
	}

	return nil
}

type EI struct {
	EntityIdentifier string `json:"1"`
	NamespaceID      string `json:"2"`
	UniversalID      string `json:"3"`
	UniversalIDType  string `json:"4"`
}

func (e *EI) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				e.EntityIdentifier = v
			case "2":
				e.NamespaceID = v
			case "3":
				e.UniversalID = v
			case "4":
				e.UniversalIDType = v
			}
		}
	}

	return nil
}

// Hierarchic Designator
type HD struct {
	NamespaceID     string `json:"1"`
	UniversalID     string `json:"2"`
	UniversalIDType string `json:"3"`
}

func (h *HD) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				h.NamespaceID = v
			case "2":
				h.UniversalID = v
			case "3":
				h.UniversalIDType = v
			}
		}
	}

	return nil
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

func (x *XCN) UnmarshalJSON(data []byte) error {
	var tempMap map[string]interface{}
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				x.IDNumber = v.(string)
			case "2":
				x.FamilyName = v.(string)
			case "3":
				x.GivenName = v.(string)
			case "4":
				x.MiddleName = v.(string)
			case "5":
				x.Suffix = v.(string)
			case "6":
				x.Prefix = v.(string)
			case "7":
				x.Degree = v.(string)
			}
		}
	}

	return nil
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

func (c *CE) UnmarshalJSON(data []byte) error {
	var tempMap map[string]string
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	for k, v := range tempMap {
		matches := re.FindStringSubmatch(k)
		if len(matches) >= 2 {
			index := matches[len(matches)-1]
			switch index {
			case "1":
				c.Identifier = v
			case "2":
				c.Text = v
			case "3":
				c.CodingSystem = v
			case "4":
				c.AltIdentifier = v
			case "5":
				c.AltText = v
			case "6":
				c.AltCodingSystem = v
			}
		}
	}

	return nil
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
