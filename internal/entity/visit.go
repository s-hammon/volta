package entity

import (
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Visit struct {
	Base
	VisitNo string
	Site    Site
	MRN     MRN
	Type    objects.PatientType
}

func DBtoVisit(visit database.GetVisitBySiteIdNumberRow) Visit {
	return Visit{
		Base: Base{
			ID:        int(visit.ID),
			CreatedAt: visit.CreatedAt.Time,
			UpdatedAt: visit.UpdatedAt.Time,
		},
		VisitNo: visit.Number,
		Site: Site{
			Base: Base{
				ID:        int(visit.SiteID.Int32),
				CreatedAt: visit.SiteCreatedAt.Time,
				UpdatedAt: visit.SiteUpdatedAt.Time,
			},
			Code:    visit.SiteCode.String,
			Name:    visit.SiteName.String,
			Address: visit.SiteAddress.String,
		},
		Type: objects.PatientType(visit.PatientType),
	}
}
