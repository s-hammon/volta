package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type orderStatus string

const (
	OrderScheduled   orderStatus = "SC"
	OrderInProgress  orderStatus = "IP"
	OrderComplete    orderStatus = "CM"
	OrderCancelled   orderStatus = "CA"
	OrderRescheduled orderStatus = "RS"
)

func newOrderStatus(status string) orderStatus {
	return orderStatus(status)
}

func (o orderStatus) String() string {
	return string(o)
}

type Order struct {
	Base
	Number        string
	CurrentStatus orderStatus
	Date          time.Time
	Site          Site
	Visit         Visit
	MRN           MRN
	Provider      Physician
}

func NewOrder(number, status string, orderDT time.Time, physician Physician) Order {
	return Order{
		Number:        number,
		CurrentStatus: newOrderStatus(status),
		Date:          orderDT,
		Provider:      physician,
	}
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
		CurrentStatus: newOrderStatus(order.CurrentStatus),
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

func (o *Order) ToDB(ctx context.Context, siteID int32, visitID, mrnID, providerID int64, db *database.Queries) (int64, string, error) {
	order, err := db.CreateOrder(ctx, database.CreateOrderParams{
		SiteID:              pgtype.Int4{Int32: siteID, Valid: true},
		VisitID:             pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:               pgtype.Int8{Int64: mrnID, Valid: true},
		OrderingPhysicianID: pgtype.Int8{Int64: providerID, Valid: true},
		Arrival:             pgtype.Timestamp{Time: o.Date, Valid: true},
		Number:              o.Number,
		CurrentStatus:       o.CurrentStatus.String(),
	})
	if err != nil {
		return 0, "", err
	}
	return order.ID, order.CurrentStatus, nil
}
