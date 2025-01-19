package entity

import (
	"github.com/google/uuid"
	"github.com/s-hammon/volta/internal/objects"
)

type Procedure struct {
	ID          uuid.UUID
	Code        string
	Description string
	Specialty   objects.Specialty
	Modality    objects.Modality
}

type FacProcedure struct {
	Code        string
	Description string
}

type FacProcedureMap map[string]FacProcedure
