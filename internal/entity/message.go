package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
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

func (m *Message) ToDB(ctx context.Context, db *database.Queries) (database.Message, error) {
	rec_dt := pgtype.Timestamp{Time: m.DateTime, Valid: true}
	res, err := db.CreateMessage(ctx, database.CreateMessageParams{
		FieldSeparator:       m.FieldSeparator,
		EncodingCharacters:   m.EncodingChars,
		SendingApplication:   m.SendingApp,
		SendingFacility:      m.SendingFac,
		ReceivingApplication: m.ReceivingApp,
		ReceivingFacility:    m.ReceivingFac,
		ReceivedAt:           rec_dt,
		MessageType:          m.Type,
		TriggerEvent:         m.TriggerEvent,
		ControlID:            m.ControlID,
		ProcessingID:         m.ProcessingID,
		VersionID:            m.Version,
	})
	if err != nil {
		return database.Message{}, err
	}

	return res, nil
}
