package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Order struct {
	Base
	Number        string
	CurrentStatus string
	Date          time.Time
	Site          Site
	Visit         Visit
	MRN           MRN
	Provider      Physician
}

func DBtoOrder(order database.GetOrderBySiteIDNumberRow) Order {
	site := Site{
		Base: Base{
			ID:        int(order.SiteID.Int32),
			CreatedAt: order.SiteCreatedAt.Time,
			UpdatedAt: order.SiteUpdatedAt.Time,
		},
		Code:    order.SiteCode.String,
		Name:    order.SiteName.String,
		Address: order.SiteAddress.String,
	}

	mrn := MRN{
		Base: Base{
			ID:        int(order.MrnID.Int64),
			CreatedAt: order.MrnCreatedAt.Time,
			UpdatedAt: order.MrnUpdatedAt.Time,
		},
		Value: order.MrnValue.String,
	}

	return Order{
		Base: Base{
			ID:        int(order.ID),
			CreatedAt: order.CreatedAt.Time,
			UpdatedAt: order.UpdatedAt.Time,
		},
		Number:        order.Number,
		CurrentStatus: order.CurrentStatus,
		Date:          order.Arrival.Time,
		Site:          site,
		Visit: Visit{
			Base: Base{
				ID:        int(order.VisitID.Int64),
				CreatedAt: order.VisitCreatedAt.Time,
				UpdatedAt: order.VisitUpdatedAt.Time,
			},
			VisitNo: order.VisitNumber.String,
			Site:    site,
			MRN:     mrn,
			Type:    objects.PatientType(order.VisitPatientType.Int16),
		},
		MRN: mrn,
		Provider: Physician{
			Base: Base{
				ID:        int(order.OrderingPhysicianID.Int64),
				CreatedAt: order.PhysicianCreatedAt.Time,
				UpdatedAt: order.PhysicianUpdatedAt.Time,
			},
			Name: objects.Name{
				First:  order.PhysicianFirstName.String,
				Last:   order.PhysicianLastName.String,
				Middle: order.PhysicianMiddleName.String,
				Suffix: order.PhysicianSuffix.String,
				Prefix: order.PhysicianPrefix.String,
				Degree: order.PhysicianDegree.String,
			},
			NPI:       order.PhysicianNpi.String,
			Specialty: objects.Specialty(order.PhysicianSpecialty.String),
		},
	}
}

func (o *Order) ToDB(ctx context.Context, siteID int32, visitID, mrnID, providerID int64, db *database.Queries) (database.Order, error) {
	return db.CreateOrder(ctx, database.CreateOrderParams{
		SiteID:              pgtype.Int4{Int32: siteID, Valid: true},
		VisitID:             pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:               pgtype.Int8{Int64: mrnID, Valid: true},
		OrderingPhysicianID: pgtype.Int8{Int64: providerID, Valid: true},
		Arrival:             pgtype.Timestamp{Time: o.Date, Valid: true},
		Number:              o.Number,
		CurrentStatus:       o.CurrentStatus,
	})
}

func (o *Order) Equal(other Order) bool {
	return o.Number == other.Number &&
		o.CurrentStatus == other.CurrentStatus &&
		o.Date == other.Date &&
		o.Site.Equal(other.Site) &&
		o.Visit.Equal(other.Visit) &&
		o.MRN.Equal(other.MRN) &&
		o.Provider.Equal(other.Provider)
}

func (o *Order) Coalesce(other Order) {
	if other.Number != "" && o.Number != other.Number {
		o.Number = other.Number
	}
	if other.CurrentStatus != "" && o.CurrentStatus != other.CurrentStatus {
		o.CurrentStatus = other.CurrentStatus
	}
	if !other.Date.IsZero() {
		o.Date = other.Date
	}
	o.Site.Coalesce(other.Site)
	o.MRN.Coalesce(other.MRN)
	o.Visit.Coalesce(other.Visit)
	o.Provider.Coalesce(other.Provider)
}
