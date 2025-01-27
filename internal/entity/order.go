package entity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
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

func (o *Order) ToDB(ctx context.Context, siteID int32, visitID, mrnID, providerID int64, db *database.Queries) (database.Order, error) {
	var arrival pgtype.Timestamp
	if err := arrival.Scan(o.Date); err != nil {
		return database.Order{}, err
	}

	res, err := db.CreateOrder(ctx, database.CreateOrderParams{
		SiteID:              pgtype.Int4{Int32: siteID, Valid: true},
		VisitID:             pgtype.Int8{Int64: visitID, Valid: true},
		MrnID:               pgtype.Int8{Int64: mrnID, Valid: true},
		OrderingPhysicianID: pgtype.Int8{Int64: providerID, Valid: true},
		Arrival:             arrival,
		Number:              o.Number,
		CurrentStatus:       o.CurrentStatus,
	})
	if err == nil {
		return res, nil
	}

	if extractErrCode(err) == "23505" {
		res, err = db.GetOrderBySiteIDNumber(ctx, database.GetOrderBySiteIDNumberParams{
			SiteID: pgtype.Int4{Int32: siteID, Valid: true},
			Number: o.Number,
		})
		if err == nil {
			return res, nil
		}
	}

	return database.Order{}, err
}
