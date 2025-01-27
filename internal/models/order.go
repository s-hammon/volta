package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type OrderModel struct {
	OrderNo          string `json:"ORC.2"`
	FillerOrderNo    string `json:"ORC.3"`
	OrderDT          string `json:"ORC.9"`
	OrderingProvider XCN    `json:"ORC.12"`
}

func (o *OrderModel) ToEntity() entity.Order {
	orderDT, err := time.Parse("20060102150405", o.OrderDT)
	if err != nil {
		orderDT = time.Now()
	}

	provider := entity.Physician{
		Name: objects.Name{
			Last:   o.OrderingProvider.FamilyName,
			First:  o.OrderingProvider.GivenName,
			Middle: o.OrderingProvider.MiddleName,
			Suffix: o.OrderingProvider.Suffix,
			Prefix: o.OrderingProvider.Prefix,
			Degree: o.OrderingProvider.Degree,
		},
		// TODO: NPI
	}

	return entity.Order{
		Number:   o.FillerOrderNo,
		Date:     orderDT,
		Provider: provider,
	}
}
