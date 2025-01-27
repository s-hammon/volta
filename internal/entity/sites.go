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

func (s *Site) ToDB(ctx context.Context, db *database.Queries) (database.Site, error) {
	res, err := db.CreateSite(ctx, database.CreateSiteParams{
		Code:    s.Code,
		Name:    s.Name,
		Address: s.Address,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetSiteByCode(ctx, s.Code)
		if err == nil {
			return res, nil
		}
	}

	return database.Site{}, err
}
