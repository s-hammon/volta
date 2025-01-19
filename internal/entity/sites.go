package entity

import "github.com/google/uuid"

type Site struct {
	ID      uuid.UUID
	Name    string
	Address string
	Phone   string
	// TODO: other fields
}
