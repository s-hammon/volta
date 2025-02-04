package models

import (
	"time"

	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type OrderModel struct {
	OrderNo          string `hl7:"ORC.2"`
	FillerOrderNo    string `hl7:"ORC.3"`
	Status           string `hl7:"ORC.5"`
	OrderDT          string `hl7:"ORC.9"`
	OrderingProvider XCN    `hl7:"ORC.12"`
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

	orderNo := o.OrderNo
	if orderNo == "" {
		orderNo = o.FillerOrderNo
	}

	return entity.Order{
		Number:        orderNo,
		CurrentStatus: o.Status,
		Date:          orderDT,
		Provider:      provider,
	}
}
