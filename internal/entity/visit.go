package entity

import "github.com/google/uuid"

type PatientType int

const (
	outPatient PatientType = iota + 1
	inPatient
	edPatient
)

type Visit struct {
	ID        uuid.UUID
	Accession string
	Site      Site
	MRN       MRN
	Type      PatientType
}
