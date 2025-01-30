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
		NPI:       physician.Npi,
		Specialty: objects.Specialty(physician.Specialty.String),
	}
}

func (p *Physician) ToDB(ctx context.Context, db *database.Queries) (database.Physician, error) {
	res, err := db.CreatePhysician(ctx, database.CreatePhysicianParams{
		FirstName:  p.Name.First,
		LastName:   p.Name.Last,
		MiddleName: pgtype.Text{String: p.Name.Middle, Valid: true},
		Suffix:     pgtype.Text{String: p.Name.Suffix, Valid: true},
		Prefix:     pgtype.Text{String: p.Name.Prefix, Valid: true},
		Degree:     pgtype.Text{String: p.Name.Degree, Valid: true},
		Npi:        p.NPI,
		Specialty:  pgtype.Text{String: p.Specialty.String(), Valid: true},
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetPhysicianByNameNPI(ctx, database.GetPhysicianByNameNPIParams{
			FirstName: p.Name.First,
			LastName:  p.Name.Last,
			Npi:       p.NPI,
		})
		if err != nil {
			return database.Physician{}, err
		}

		ph := DBtoPhysician(res)
		if !ph.Equal(*p) {
			ph.Coalesce(*p)
			return db.UpdatePhysician(ctx, database.UpdatePhysicianParams{
				ID:         int64(ph.ID),
				FirstName:  ph.Name.First,
				LastName:   ph.Name.Last,
				MiddleName: pgtype.Text{String: ph.Name.Middle, Valid: true},
				Suffix:     pgtype.Text{String: ph.Name.Suffix, Valid: true},
				Prefix:     pgtype.Text{String: ph.Name.Prefix, Valid: true},
				Degree:     pgtype.Text{String: ph.Name.Degree, Valid: true},
				Npi:        ph.NPI,
				Specialty:  pgtype.Text{String: ph.Specialty.String(), Valid: true},
			})
		}
	}

	return database.Physician{}, err
}

func (p *Physician) Equal(other Physician) bool {
	return p.Name.Full() == other.Name.Full() &&
		p.NPI == other.NPI &&
		p.Specialty == other.Specialty
}

func (p *Physician) Coalesce(other Physician) {
	p.Name.Coalesce(other.Name)
	if other.NPI != "" && p.NPI != other.NPI {
		p.NPI = other.NPI
	}
	if other.Specialty != "" && p.Specialty != other.Specialty {
		p.Specialty = other.Specialty
	}
}
