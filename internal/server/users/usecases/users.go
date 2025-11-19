package usecases

import (
	"context"

	"passman/internal/server/users"

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
	if existedUser, err := uu.dbRepo.GetUser(ctx, userCreds.Username); err != nil && !uu.dbRepo.IsEmptyRows(err) {
		return uuid.UUID{}, newInternalError("Registration", "failed finding user", err)
	} else if len(existedUser.Username) > 0 {
		return uuid.UUID{}, errUserExist
	}

	hash, err := argon2id.CreateHash(userCreds.Password, argon2id.DefaultParams)
	if err != nil {
		return uuid.UUID{}, newInternalError("Registration", "failed creating password hash", err)
	}

	newUser := users.User{
		ID:       uuid.New(),
		Username: userCreds.Username,
		Password: hash,
	}

	if err := uu.dbRepo.AddUser(ctx, newUser); err != nil {
		return uuid.UUID{}, newInternalError("Registration", "failed adding user to db", err)
	}

	return newUser.ID, nil
}

func (uu *userUsecase) Login(ctx context.Context, userCreds users.UserDTO) (uuid.UUID, error) {
	user, err := uu.dbRepo.GetUser(ctx, userCreds.Username)
	if err != nil {
		if uu.dbRepo.IsEmptyRows(err) {
			return uuid.UUID{}, errUserNotFound
		}
		return uuid.UUID{}, newInternalError("Login", "failed finding user", err)
	}

	match, err := argon2id.ComparePasswordAndHash(userCreds.Password, user.Password)
	if err != nil {
		return uuid.UUID{}, newInternalError("Login", "failed comparing password", err)
	}
	if !match {
		return uuid.UUID{}, errIncorrectPassword
	}

	return user.ID, nil
}

func (uu *userUsecase) UpdateUser(ctx context.Context, updatedParameters users.UpdatedUserParams) error {
	user, err := uu.dbRepo.GetUserByID(ctx, updatedParameters.UserID)
	if err != nil {
		return newInternalError("UpdateUser", "failed finding user", err)
	}

	match, err := argon2id.ComparePasswordAndHash(updatedParameters.OldPassword, user.Password)
	if err != nil {
		return newInternalError("UpdateUser", "failed comparing password", err)
	}
	if !match {
		return errIncorrectPassword
	}

	if len(updatedParameters.Username) == 0 {
		updatedParameters.Username = user.Username
	}

	if len(updatedParameters.Password) > 0 {
		hash, err := argon2id.CreateHash(updatedParameters.Password, argon2id.DefaultParams)
		if err != nil {
			return newInternalError("UpdateUser", "failed creating password hash", err)
		}
		updatedParameters.Password = hash
	} else {
		updatedParameters.Password = user.Password
	}

	updatedUser := users.User{
		ID:       updatedParameters.UserID,
		Username: updatedParameters.Username,
		Password: updatedParameters.Password,
	}

	if err := uu.dbRepo.UpdateUser(ctx, updatedUser); err != nil {
		return newInternalError("UpdateUser", "failed updating user", err)
	}

	return nil
}

func (uu *userUsecase) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	if err := uu.dbRepo.RemoveUser(ctx, userID); err != nil {
		return newInternalError("DeleteUser", "failed removing user", err)
	}
	return nil
}

func (uu *userUsecase) ParseUserError(err error) (int, string, error) {
	return parseError(err)
}
