package models

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/s-hammon/volta/internal/database"
)

type ORM struct {
	MSH MessageModel `json:"MSH"`
	PID PatientModel `json:"PID"`
	PV1 VisitModel   `json:"PV1"`
	ORC OrderModel   `json:"ORC"`
	OBR ExamModel    `json:"OBR"`
}

func (orm *ORM) ToDB(ctx context.Context, db *database.Queries) error {
	p := orm.PID.ToEntity()
	m := orm.MSH.ToEntity()
	v := orm.PV1.ToEntity(m.SendingFac, orm.PID.MRN)
	o := orm.ORC.ToEntity()
	e := orm.OBR.ToEntity(v.Site.Code, o.CurrentStatus.String(), orm.PID.MRN)

	_, err := m.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating message: " + err.Error())
	}

	siteID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating site: " + err.Error())
	}

	patientID, err := p.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating patient: " + err.Error())
	}

	mrnID, err := v.MRN.ToDB(ctx, siteID, patientID, db)
	if err != nil {
		return errors.New("error creating mrn: " + err.Error())
	}

	if v.VisitNo == "" {
		// set this equal to the order number--it's the best we can do :/
		log.Debug().Str("orderNumber", o.Number).Msg("filling visit number with order number")
		v.VisitNo = o.Number
	}
	visitID, err := v.ToDB(ctx, siteID, mrnID, db)
	if err != nil {
		return errors.New("error creating visit: " + err.Error())
	}

	physicianID, err := o.Provider.ToDB(ctx, db)
	if err != nil {
		return errors.New("error creating physician: " + err.Error())
	}

	orderID, orderStatus, err := o.ToDB(ctx, siteID, visitID, mrnID, physicianID, db)
	if err != nil {
		return errors.New("error creating order: " + err.Error())
	}

	procedureID, err := e.Procedure.ToDB(ctx, siteID, db)
	if err != nil {
		return errors.New("error creating procedure: " + err.Error())
	}

	_, err = e.ToDB(ctx, orderID, visitID, mrnID, siteID, procedureID, orderStatus, db)
	if err != nil {
		return errors.New("error creating exam: " + err.Error())
	}

	return nil
}
