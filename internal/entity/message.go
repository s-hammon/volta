package entity

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID
	FieldSeparator string
	EncodingChars  string
	SendingApp     string
	SendingFac     string
	ReceivingApp   string
	ReceivingFac   string
	DateTime       time.Time
	Type           string
	TriggerEvent   string
	ControlID      string
	ProcessingID   string
	Version        string
}
