package usecases

import (
	"context"
	"errors"
	"testing"

	"passman/internal/server/users"
	mock_usecases "passman/internal/server/users/usecases/mock"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockdbRepo(ctrl)
	userUsecase := New(mockRepo)
	ctx := context.Background()
	errEmptyRows := errors.New("empty rows")
	userCreds := users.UserDTO{
		Username: "test_user",
		Password: "test_password",
	}

	type findUserResult struct {
		existedUser users.User
		err         error
	}

	type addUserResult struct {
		err error
	}

	type expResult struct {
		userID uuid.UUID
		err    error
	}

	tests := []struct {
		name           string
		findUserResult *findUserResult
		addUserResult  *addUserResult
		expResult      expResult
	}{
		{
			name:           "find_user_error",
			findUserResult: &findUserResult{err: errors.New("internal error")},
			expResult:      expResult{err: errors.New("Registration: failed finding user")},
		},
		{
			name:           "user_already_exist",
			findUserResult: &findUserResult{existedUser: users.User{Username: "founded_user"}},
			expResult:      expResult{err: errors.New("ClientError: user already exist")},
		},
		{
			name:           "add_user_error",
			findUserResult: &findUserResult{err: errEmptyRows},
			addUserResult:  &addUserResult{err: errors.New("internal error")},
			expResult:      expResult{err: errors.New("Registration: failed adding user to db")},
		},
		{
			name:           "success",
			findUserResult: &findUserResult{err: errEmptyRows},
			addUserResult:  &addUserResult{err: nil},
			expResult:      expResult{userID: uuid.New()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUser(ctx, userCreds.Username).
				Return(test.findUserResult.existedUser, test.findUserResult.err).
				Times(1)

			if test.findUserResult.err != nil {
				isEmpty := mockRepo.EXPECT().IsEmptyRows(test.findUserResult.err)
				if test.findUserResult.err == errEmptyRows {
					isEmpty.Return(true).Times(1)
				} else {
					isEmpty.Return(false).Times(1)
				}
			}

			if test.addUserResult != nil {
				mockRepo.EXPECT().
					AddUser(ctx, gomock.AssignableToTypeOf(users.User{})).
					Return(test.addUserResult.err).
					Times(1)
			}

			_, err := userUsecase.Registration(ctx, userCreds)

			if got, want := err, test.expResult.err; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %d\n\tActual: %d", want, got)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockdbRepo(ctrl)
	userUsecase := New(mockRepo)
	ctx := context.Background()
	errEmptyRows := errors.New("empty rows")
	userCreds := users.UserDTO{
		Username: "test_user",
		Password: "test_password",
	}
	hashUserPassword, _ := argon2id.CreateHash(userCreds.Password, argon2id.DefaultParams)
	foundedUser := users.User{
		ID:       uuid.New(),
		Username: "tetst_user",
		Password: hashUserPassword,
	}

	type findUserResult struct {
		existedUser users.User
		err         error
	}

	type expResult struct {
		userID uuid.UUID
		err    error
	}

	tests := []struct {
		name           string
		findUserResult *findUserResult
		expResult      expResult
	}{
		{
			name:           "find_user_error",
			findUserResult: &findUserResult{err: errors.New("internal error")},
			expResult:      expResult{err: errors.New("Login: failed finding user")},
		},
		{
			name:           "user_not_found",
			findUserResult: &findUserResult{err: errEmptyRows},
			expResult:      expResult{err: errors.New("ClientError: user not found")},
		},
		{
			name:           "password_is_not_a_hash",
			findUserResult: &findUserResult{existedUser: users.User{Password: "not a hash"}},
			expResult:      expResult{err: errors.New("Login: failed comparing password")},
		},
		{
			name:           "incorrect_password",
			findUserResult: &findUserResult{existedUser: users.User{Password: "$argon2id$v=19$m=65536,t=1,p=4$deAxNTdiK57uVLnhNR+FqA$xltWpE8oWxA9nifflVJOtdXvsVXhgU13oabfsdv/GeY"}},
			expResult:      expResult{err: errors.New("ClientError: incorrect password")},
		},
		{
			name:           "success",
			findUserResult: &findUserResult{existedUser: foundedUser},
			expResult:      expResult{userID: foundedUser.ID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUser(ctx, userCreds.Username).
				Return(test.findUserResult.existedUser, test.findUserResult.err).
				Times(1)

			if test.findUserResult.err != nil {
				isEmpty := mockRepo.EXPECT().IsEmptyRows(test.findUserResult.err)
				if test.findUserResult.err == errEmptyRows {
					isEmpty.Return(true).Times(1)
				} else {
					isEmpty.Return(false).Times(1)
				}
			}

			userID, err := userUsecase.Login(ctx, userCreds)

			if got, want := err, test.expResult.err; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %d\n\tActual: %d", want, got)
			}

			if got, want := userID, test.expResult.userID; got != want {
				t.Fatalf("Wrong! Unexpected userID!\n\tExpected: %s\n\tActual: %s", want.String(), got.String())
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockdbRepo(ctrl)
	userUsecase := New(mockRepo)

	ctx := context.Background()
	userFromDB := users.User{
		Username: "user",
		Password: "$argon2id$v=19$m=65536,t=1,p=4$XAPvqtpAVs/NGpyd1H5Fmg$pvbpnBLwlbXfFyuRmochGwJwetm1rv2m1/MmCw7qcPc",
	}

	incorrectParams := users.UpdatedUserParams{
		UserID:      uuid.New(),
		OldPassword: "incorrect_new_password",
	}

	correctParams := users.UpdatedUserParams{
		UserID:      uuid.New(),
		OldPassword: "user_password",
		UserDTO: users.UserDTO{
			Username: "some_user",
			Password: "some_password",
		},
	}

	type getUserResult struct {
		user users.User
		err  error
	}

	type updateUserResult struct {
		err error
	}

	tests := []struct {
		name             string
		input            users.UpdatedUserParams
		getUserResult    *getUserResult
		updateUserResult *updateUserResult
		expResult        error
	}{
		{
			name:          "failed_finding_user",
			getUserResult: &getUserResult{err: errors.New("internal error")},
			expResult:     errors.New("UpdateUser: failed finding user"),
		},
		{
			name:          "incorrect_password",
			input:         incorrectParams,
			getUserResult: &getUserResult{user: userFromDB},
			expResult:     errors.New("ClientError: incorrect password"),
		},
		{
			name:             "failed_updating_user",
			input:            correctParams,
			getUserResult:    &getUserResult{user: userFromDB},
			updateUserResult: &updateUserResult{err: errors.New("internal error")},
			expResult:        errors.New("UpdateUser: failed updating user"),
		},
		{
			name:             "success",
			input:            correctParams,
			getUserResult:    &getUserResult{user: userFromDB},
			updateUserResult: &updateUserResult{err: nil},
			expResult:        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUserByID(ctx, test.input.UserID).
				Return(test.getUserResult.user, test.getUserResult.err).
				Times(1)

			if test.updateUserResult != nil {
				mockRepo.EXPECT().
					UpdateUser(ctx, gomock.AssignableToTypeOf(users.User{})).
					Return(test.updateUserResult.err).
					Times(1)
			}

			actErr := userUsecase.UpdateUser(ctx, test.input)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %d\n\tActual: %d", want, got)
			}
		})
	}
}
