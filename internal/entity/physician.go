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
		if err == nil {
			return res, nil
		}
	}

	return database.Physician{}, err
}
