package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
)

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
	dt, err := time.Parse("20060102150405", m.DateTime)
	if err != nil {
		dt = time.Now()
	}
	return entity.Message{
		FieldSeparator: m.FieldSeparator,
		EncodingChars:  m.EncodingChars,
		SendingApp:     m.SendingApp,
		SendingFac:     m.SendingFac,
		ReceivingApp:   m.ReceivingApp,
		ReceivingFac:   m.ReceivingFac,
		DateTime:       dt,
		Type:           m.Type.Type,
		TriggerEvent:   m.Type.TriggerEvent,
		ControlID:      m.ControlID,
		ProcessingID:   m.ProcessingID,
		Version:        m.Version,
	}
}
