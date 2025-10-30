package usecases

import (
	"context"
	"errors"
	"passman/cmd/internal/users"
	mock_usecases "passman/cmd/internal/users/usecases/mock"
	"testing"

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
			expResult:      expResult{err: errors.New("Registration: can't find user")},
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
			expResult:      expResult{err: errors.New("Registration: can't add user to db")},
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
				FindUser(ctx, userCreds.Username).
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
			expResult:      expResult{err: errors.New("Login: can't find user")},
		},
		{
			name:           "user_not_found",
			findUserResult: &findUserResult{err: errEmptyRows},
			expResult:      expResult{err: errors.New("ClientError: user not found")},
		},
		{
			name:           "password_is_not_a_hash",
			findUserResult: &findUserResult{existedUser: users.User{Password: "not a hash"}},
			expResult:      expResult{err: errors.New("Login: can't compare password")},
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
				FindUser(ctx, userCreds.Username).
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
