package api

import (
	"context"
	"errors"

	"cloud.google.com/go/pubsub"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

const segDelim = "\r"

type HealthcareService interface {
	GetHL7V2Message(messagePath string) ([]byte, error)
}

type API struct {
	DB  *database.Queries
	Svc HealthcareService
}

type Handler func(context.Context, *pubsub.Message) (models.Response, error)

func (a *API) HandleORM(ctx context.Context, msg *pubsub.Message) (models.Response, error) {
	var resp models.Response
	raw, err := a.Svc.GetHL7V2Message(string(msg.Data))
	if err != nil {
		return resp, errors.New("error getting HL7 message" + err.Error())
	}

	msgMap, err := hl7.NewMessage(raw, []byte(segDelim))
	if err != nil {
		return resp, errors.New("error creating message from raw: " + err.Error())
	}

	orm, err := models.NewORM(msgMap)
	if err != nil {
		return resp, errors.New("error creating ORM: " + err.Error())
	}

	resp, err = orm.ToDB(ctx, a.DB)
	if err != nil {
		return resp, errors.New("error processing ORM: " + err.Error())
	}

	return resp, nil
}
