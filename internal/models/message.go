package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
)

type MessageModel struct {
	FieldSeparator string `json:"MSH.1"`
	EncodingChars  string `json:"MSH.2"`
	SendingApp     string `json:"MSH.3"`
	SendingFac     string `json:"MSH.4"`
	ReceivingApp   string `json:"MSH.5"`
	ReceivingFac   string `json:"MSH.6"`
	DateTime       string `json:"MSH.7"`
	Type           CM_MSG `json:"MSH.9"`
	ControlID      string `json:"MSH.10"`
	ProcessingID   string `json:"MSH.11"`
	Version        string `json:"MSH.12"`
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
