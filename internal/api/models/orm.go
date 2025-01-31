package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/s-hammon/volta/internal/database"
)

type Response struct {
	Message  string `json:"message"`
	Entities []byte `json:"entities"`
}

type ORM struct {
	MSH MessageModel `json:"MSH"`
	PID PatientModel `json:"PID"`
	PV1 VisitModel   `json:"PV1"`
	ORC OrderModel   `json:"ORC"`
	OBR ExamModel    `json:"OBR"`
}

func NewORM(msgMap map[string]interface{}) (ORM, error) {
	b, err := json.Marshal(msgMap)
	if err != nil {
		return ORM{}, err
	}

	orm := ORM{}
	if err = json.Unmarshal(b, &orm); err != nil {
		return ORM{}, err
	}

	return orm, nil
}

func (orm *ORM) ToDB(ctx context.Context, db *database.Queries) (Response, error) {
	var r Response
	entities := map[string]interface{}{}

	p := orm.PID.ToEntity()
	m := orm.MSH.ToEntity()
	v := orm.PV1.ToEntity(m.SendingFac, orm.PID.MRN)
	o := orm.ORC.ToEntity()
	e := orm.OBR.ToEntity(v.Site.Code, o.CurrentStatus, orm.PID.MRN)

	site, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating site: "+err.Error(), r, entities)
	}
	entities["site"] = site

	patient, err := p.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating patient: "+err.Error(), r, entities)
	}
	entities["patient"] = patient

	mrn, err := v.MRN.ToDB(ctx, site.ID, patient.ID, db)
	if err != nil {
		return handleError("error creating mrn: "+err.Error(), r, entities)
	}
	entities["mrn"] = mrn

	if v.VisitNo == "" {
		// set this equal to the order number--it's the best we can do :/
		fmt.Printf("filling visit number with order number: %s\n", o.Number)
		v.VisitNo = o.Number
	}
	visit, err := v.ToDB(ctx, site.ID, mrn.ID, db)
	if err != nil {
		return handleError("error creating visit: "+err.Error(), r, entities)
	}
	entities["visit"] = visit

	physician, err := o.Provider.ToDB(ctx, db)
	if err != nil {
		return handleError("error creating physician: "+err.Error(), r, entities)
	}
	entities["physician"] = physician

	order, err := o.ToDB(ctx, site.ID, visit.ID, mrn.ID, physician.ID, db)
	if err != nil {
		return handleError("error creating order: "+err.Error(), r, entities)
	}
	entities["order"] = order

	procedure, err := e.Procedure.ToDB(ctx, site.ID, db)
	if err != nil {
		return handleError("error creating procedure: "+err.Error(), r, entities)
	}
	entities["procedure"] = procedure

	exam, err := e.ToDB(ctx, order.ID, visit.ID, mrn.ID, site.ID, procedure.ID, order.CurrentStatus, db)
	if err != nil {
		return handleError("error creating exam: "+err.Error(), r, entities)
	}
	entities["exam"] = exam

	b, err := json.Marshal(entities)
	if err != nil {
		return handleError("error marshalling entities: "+err.Error(), r, entities)
	}
	r.Entities = b

	r.Message = "success"

	return r, nil
}

func handleError(errMsg string, r Response, entities map[string]interface{}) (Response, error) {
	r.Message = errMsg

	b, err := json.Marshal(entities)
	if err != nil {
		return r, err
	}
	r.Entities = b

	return r, errors.New(errMsg)
}
