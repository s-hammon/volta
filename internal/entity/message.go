package entity

import (
	"time"
)

type Message struct {
	Base
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
