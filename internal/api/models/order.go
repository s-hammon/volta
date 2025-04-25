package models

import (
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/internal/objects"
)

type OrderModel struct {
	OrderNo          string `json:"ORC.2"`
	FillerOrderNo    string `json:"ORC.3"`
	Status           string `json:"ORC.5"`
	OrderDT          string `json:"ORC.9"`
	OrderingProvider XCN    `json:"ORC.12"`
}

func (o *OrderModel) ToEntity() entity.Order {
	orderDT := convertCSTtoUTC(o.OrderDT)
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

	return entity.NewOrder(orderNo, o.Status, orderDT, provider)
}
