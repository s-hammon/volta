package entity

import "github.com/s-hammon/volta/internal/objects"

type Procedure struct {
	Base
	Site        Site
	Code        string
	Description string
	Specialty   objects.Specialty
	Modality    objects.Modality
}
