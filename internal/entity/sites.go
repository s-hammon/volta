package entity

import (
	"context"

	"github.com/s-hammon/volta/internal/database"
)

type Site struct {
	Base
	Code    string
	Name    string
	Address string
	Phone   string
	// TODO: other fields
}

func DBtoSite(site database.Site) Site {
	return Site{
		Base: Base{
			ID:        int(site.ID),
			CreatedAt: site.CreatedAt.Time,
			UpdatedAt: site.UpdatedAt.Time,
		},
		Code:    site.Code,
		Name:    site.Name,
		Address: site.Address,
		// TODO: handle phone numbers
	}
}

func (s *Site) ToDB(ctx context.Context, db *database.Queries) (int32, error) {
	res, err := db.CreateSite(ctx, database.CreateSiteParams{
		Code:    s.Code,
		Name:    s.Name,
		Address: s.Address,
	})
	if err != nil {
		return 0, err
	}
	return res.ID, nil
}

func (s *Site) Equal(other Site) bool {
	return s.Code == other.Code &&
		s.Name == other.Name &&
		s.Address == other.Address &&
		s.Phone == other.Phone
}

func (s *Site) Coalesce(other Site) {
	if other.Code != "" && s.Code != other.Code {
		s.Code = other.Code
	}
	if other.Name != "" && s.Name != other.Name {
		s.Name = other.Name
	}
	if other.Address != "" && s.Address != other.Address {
		s.Address = other.Address
	}
	if other.Phone != "" && s.Phone != other.Phone {
		s.Phone = other.Phone
	}
}
