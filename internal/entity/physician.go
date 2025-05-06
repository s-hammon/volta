package entity

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
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
		AppCode:   physician.AppCode.String,
		NPI:       physician.Npi,
		Specialty: objects.Specialty(physician.Specialty.String),
	}
}

func (p *Physician) ToDB(ctx context.Context, db *database.Queries) (int64, error) {
	phys, err := db.CreatePhysician(ctx, database.CreatePhysicianParams{
		FirstName:  p.Name.First,
		LastName:   p.Name.Last,
		MiddleName: pgtype.Text{String: p.Name.Middle, Valid: true},
		Suffix:     pgtype.Text{String: p.Name.Suffix, Valid: true},
		Prefix:     pgtype.Text{String: p.Name.Prefix, Valid: true},
		Degree:     pgtype.Text{String: p.Name.Degree, Valid: true},
		AppCode:    pgtype.Text{String: p.AppCode, Valid: true},
		Npi:        p.NPI,
	})
	if err != nil {
		return 0, err
	}
	return phys.ID, nil
}
