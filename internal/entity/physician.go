package entity

import (
	"github.com/google/uuid"
	"github.com/s-hammon/volta/internal/objects"
)

type Physician struct {
	ID        uuid.UUID
	Name      objects.Name
	NPI       string
	Specialty objects.Specialty
}
