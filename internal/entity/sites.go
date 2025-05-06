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
		Name:    site.Name.String,
		Address: site.Address,
		// TODO: handle phone numbers
	}
}

func (s *Site) ToDB(ctx context.Context, db *database.Queries) (int32, error) {
	params := database.CreateSiteParams{
		Code:    s.Code,
		Address: s.Address,
	}
	if s.Name != "" {
		params.Name.String = s.Name
	}
	res, err := db.CreateSite(ctx, params)
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
