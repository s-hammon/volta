package entity

import "github.com/s-hammon/volta/internal/objects"

type Procedure struct {
	Base
	Site        Site              `json:"site,omitempty"`
	Code        string            `json:"code"`
	Description string            `json:"description"`
	Specialty   objects.Specialty `json:"specialty,omitempty"`
	Modality    objects.Modality  `json:"modality,omitempty"`
}
