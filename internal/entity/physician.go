package entity

import (
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Physician struct {
	Base
	Name      objects.Name
	AppCode   string
	NPI       string
	Specialty objects.Specialty
}

func DBtoPhysician(physician database.Physician) Physician {
	return Physician{
		Base: Base{
			ID:        int(physician.ID),
			CreatedAt: physician.CreatedAt.Time,
			UpdatedAt: physician.UpdatedAt.Time,
		},
		Name: objects.Name{
			First:  physician.FirstName,
			Last:   physician.LastName,
			Middle: physician.MiddleName.String,
			Suffix: physician.Suffix.String,
			Prefix: physician.Prefix.String,
			Degree: physician.Degree.String,
		},
		AppCode:   physician.AppCode,
		NPI:       physician.Npi.String,
		Specialty: objects.Specialty(physician.Specialty.String),
	}
}
