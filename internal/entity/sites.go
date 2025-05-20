package entity

import "github.com/s-hammon/volta/internal/database"

type Site struct {
	Base
	Code    string `json:"code"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
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

func (s *Site) Equal(other Site) bool {
	return s.Code == other.Code &&
		s.Name == other.Name &&
		s.Address == other.Address &&
		s.Phone == other.Phone
}
