package users

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Username string
	Password string
}

type UserDTO struct {
	Username string
	Password string
}
