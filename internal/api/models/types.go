package models

import (
	"encoding/json"
	"strings"
)

// Messate Type
type CM_MSG struct {
	Type         string `json:"1"`
	TriggerEvent string `json:"2"`
}

func (m *CM_MSG) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			m.Type = v
		case strings.HasSuffix(k, ".2"):
			m.TriggerEvent = v
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

func (x *XPN) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			x.LastName = v
		case strings.HasSuffix(k, ".2"):
			x.FirstName = v
		case strings.HasSuffix(k, ".3"):
			x.MiddleName = v
		case strings.HasSuffix(k, ".4"):
			x.Suffix = v
		case strings.HasSuffix(k, ".5"):
			x.Prefix = v
		case strings.HasSuffix(k, ".6"):
			x.Degree = v
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

func (c *CX) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		c.ID = string(data)
		return nil
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			c.ID = v
		case strings.HasSuffix(k, ".2"):
			c.CheckDigit = v
		case strings.HasSuffix(k, ".3"):
			c.CheckDigitScheme = v
		case strings.HasSuffix(k, ".4"):
			c.AssigningAuthority = v
		case strings.HasSuffix(k, ".5"):
			c.IdentifierTypeCode = v
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
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			p.PointOfCare = v
		case strings.HasSuffix(k, ".2"):
			p.Room = v
		case strings.HasSuffix(k, ".3"):
			p.Bed = v
		case strings.HasSuffix(k, ".4"):
			p.Facility = v
		case strings.HasSuffix(k, ".5"):
			p.LocationStatus = v
		case strings.HasSuffix(k, ".6"):
			p.PersonLocationType = v
		case strings.HasSuffix(k, ".7"):
			p.Building = v
		case strings.HasSuffix(k, ".8"):
			p.Floor = v
		case strings.HasSuffix(k, ".9"):
			p.LocationDescription = v
		}
	}
	return nil
}

// Entity Identifier
type EI struct {
	EntityIdentifier string `json:"1"`
	NamespaceID      string `json:"2"`
	UniversalID      string `json:"3"`
	UniversalIDType  string `json:"4"`
}

func (e *EI) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			e.EntityIdentifier = v
		case strings.HasSuffix(k, ".2"):
			e.NamespaceID = v
		case strings.HasSuffix(k, ".3"):
			e.UniversalID = v
		case strings.HasSuffix(k, ".4"):
			e.UniversalIDType = v
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
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			h.NamespaceID = v
		case strings.HasSuffix(k, ".2"):
			h.UniversalID = v
		case strings.HasSuffix(k, ".3"):
			h.UniversalIDType = v
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
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			x.IDNumber = string(v)
		case strings.HasSuffix(k, ".2"):
			x.FamilyName = string(v)
		case strings.HasSuffix(k, ".3"):
			x.GivenName = string(v)
		case strings.HasSuffix(k, ".4"):
			x.MiddleName = string(v)
		case strings.HasSuffix(k, ".5"):
			x.Suffix = string(v)
		case strings.HasSuffix(k, ".6"):
			x.Prefix = string(v)
		case strings.HasSuffix(k, ".7"):
			x.Degree = string(v)
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
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			c.Identifier = v
		case strings.HasSuffix(k, ".2"):
			c.Text = v
		case strings.HasSuffix(k, ".3"):
			c.CodingSystem = v
		case strings.HasSuffix(k, ".4"):
			c.AltIdentifier = v
		case strings.HasSuffix(k, ".5"):
			c.AltText = v
		case strings.HasSuffix(k, ".6"):
			c.AltCodingSystem = v
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

func (x *XAD) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			x.StreetAddress = v
		case strings.HasSuffix(k, ".2"):
			x.OtherDesignation = v
		case strings.HasSuffix(k, ".3"):
			x.City = v
		case strings.HasSuffix(k, ".4"):
			x.State = v
		case strings.HasSuffix(k, ".5"):
			x.Zip = v
		case strings.HasSuffix(k, ".6"):
			x.Country = v
		}
	}
	return nil
}

// Timing Quantity
type TQ struct {
	Quantity string `json:"1"`
	Interval string `json:"2"`
	Duration string `json:"3"`
	StartDT  string `json:"4"`
	EndDT    string `json:"5"`
}

func (t *TQ) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch {
		case strings.HasSuffix(k, ".1"):
			t.Quantity = v
		case strings.HasSuffix(k, ".2"):
			t.Interval = v
		case strings.HasSuffix(k, ".3"):
			t.Duration = v
		case strings.HasSuffix(k, ".4"):
			t.StartDT = v
		case strings.HasSuffix(k, ".5"):
			t.EndDT = v
		}
	}
	return nil
}
