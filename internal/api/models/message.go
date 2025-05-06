package models

import "github.com/s-hammon/volta/internal/entity"

type MessageModel struct {
	FieldSeparator string `hl7:"MSH.1"`
	EncodingChars  string `hl7:"MSH.2"`
	SendingApp     string `hl7:"MSH.3"`
	SendingFac     string `hl7:"MSH.4"`
	ReceivingApp   string `hl7:"MSH.5"`
	ReceivingFac   string `hl7:"MSH.6"`
	DateTime       string `hl7:"MSH.7"`
	Type           CM_MSG `hl7:"MSH.9"`
	ControlID      string `hl7:"MSH.10"`
	ProcessingID   string `hl7:"MSH.11"`
	Version        string `hl7:"MSH.12"`
}

func (m *MessageModel) ToEntity() entity.Message {
	dt := convertCSTtoUTC(m.DateTime)
	return entity.Message{
		FieldSeparator: m.FieldSeparator,
		EncodingChars:  m.EncodingChars,
		SendingApp:     m.SendingApp,
		SendingFac:     m.SendingFac,
		ReceivingApp:   m.ReceivingApp,
		ReceivingFac:   m.ReceivingFac,
		DateTime:       dt,
		Type:           m.Type.Name,
		TriggerEvent:   m.Type.TriggerEvent,
		ControlID:      m.ControlID,
		ProcessingID:   m.ProcessingID,
		Version:        m.Version,
	}
}
