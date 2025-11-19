package services

import "github.com/google/uuid"

type Service struct {
	ID   uuid.UUID
	Name string
	Logo string
}

type ServiceDTO struct {
	Name string
	Logo string
}
