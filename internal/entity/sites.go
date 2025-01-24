package entity

import "github.com/google/uuid"

type Site struct {
	ID      uuid.UUID
	Code    string
	Name    string
	Address string
	Phone   string
	// TODO: other fields
}
