package main

import (
	"fmt"
	"log"
	"os"

	"github.com/s-hammon/volta/internal/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

func main() {
	b, err := os.ReadFile("orm_02.txt")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	msg, err := hl7.NewMessage(b, []byte("\r"))
	if err != nil {
		log.Fatalf("error creating message: %v", err)
	}

	orm, err := models.NewORM(msg)
	if err != nil {
		log.Fatalf("error creating orm: %v", err)
	}

	patient := orm.PID.ToEntity()
	fmt.Println(patient.String())
}
