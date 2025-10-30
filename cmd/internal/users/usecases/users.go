package usecases

import (
	"context"
	"passman/cmd/internal/users"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

var (
	errUserExist         = newClientError("user already exist")
	errUserNotFound      = newClientError("user not found")
	errIncorrectPassword = newClientError("incorrect password")
)

type userUsecase struct {
	dbRepo dbRepo
}

func New(db dbRepo) *userUsecase {
	return &userUsecase{dbRepo: db}
}

func (uu *userUsecase) Registration(ctx context.Context, userCreds users.UserDTO) (uuid.UUID, error) {
	if existedUser, err := uu.dbRepo.FindUser(ctx, userCreds.Username); err != nil && !uu.dbRepo.IsEmptyRows(err) {
		return uuid.UUID{}, newInternalError("Registration", "can't find user", err)
	} else if len(existedUser.Username) > 0 {
		return uuid.UUID{}, errUserExist
	}

	hash, err := argon2id.CreateHash(userCreds.Password, argon2id.DefaultParams)
	if err != nil {
		return uuid.UUID{}, newInternalError("Registration", "can't create password hash", err)
	}

	newUser := users.User{
		ID:       uuid.New(),
		Username: userCreds.Username,
		Password: hash,
	}

	if err := uu.dbRepo.AddUser(ctx, newUser); err != nil {
		return uuid.UUID{}, newInternalError("Registration", "can't add user to db", err)
	}

	return newUser.ID, nil
}

func (uu *userUsecase) Login(ctx context.Context, userCreds users.UserDTO) (uuid.UUID, error) {
	user, err := uu.dbRepo.FindUser(ctx, userCreds.Username)
	if err != nil {
		if uu.dbRepo.IsEmptyRows(err) {
			return uuid.UUID{}, errUserNotFound
		}
		return uuid.UUID{}, newInternalError("Login", "can't find user", err)
	}

	match, err := argon2id.ComparePasswordAndHash(userCreds.Password, user.Password)
	if err != nil {
		return uuid.UUID{}, newInternalError("Login", "can't compare password", err)
	}
	if !match {
		return uuid.UUID{}, errIncorrectPassword
	}

	return user.ID, nil
}

func (uu *userUsecase) UpdateUser(ctx context.Context, userID uuid.UUID, updatedUserDTO users.UserDTO) error {
	hash, err := argon2id.CreateHash(updatedUserDTO.Password, argon2id.DefaultParams)
	if err != nil {
		return newInternalError("UpdateUser", "can't create password hash", err)
	}

	updatedUser := users.User{
		ID:       userID,
		Username: updatedUserDTO.Username,
		Password: hash,
	}

	if err := uu.dbRepo.UpdateUser(ctx, updatedUser); err != nil {
		return newInternalError("UpdateUser", "can't update user", err)
	}

	return nil
}

func (uu *userUsecase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := uu.dbRepo.DeleteUser(ctx, userID); err != nil {
		return newInternalError("DeleteUser", "can't delete user", err)
	}
	return nil
}

func (uu *userUsecase) ParseUserError(err error) (int, string, error) {
	return parseError(err)
}
