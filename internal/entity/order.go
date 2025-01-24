package entity

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID
	Accession string
	Date      time.Time
	Provider  Physician
}
