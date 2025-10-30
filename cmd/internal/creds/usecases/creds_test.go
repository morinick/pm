package usecases

import (
	"context"
	"encoding/hex"
	"errors"
	"passman/cmd/internal/creds"
	mock_usecases "passman/cmd/internal/creds/usecases/mock"
	"passman/pkg/cipher"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func generateTestCiphers() []cipher.AESCipher {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	key, _ := hex.DecodeString(hexKey)
	ciph, _ := cipher.New(key)
	return []cipher.AESCipher{*ciph}
}

func TestAddNewCreds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := mock_usecases.NewMockrepository(ctrl)
	ciphers := generateTestCiphers()
	usecase := New(mockRepository, ciphers)

	userID := uuid.New()
	serviceCreds := creds.Service{
		Name:     "test_service_name",
		Login:    "test_service_login",
		Password: "test_service_password",
	}
	errNoRows := errors.New("no rows")
	ctx := context.Background()

	type findCredsResult struct {
		existedCreds creds.ServiceDAO
		err          error
	}

	type addNewCredsResult struct {
		err error
	}

	tests := []struct {
		name              string
		findCredsResult   *findCredsResult
		addNewCredsResult *addNewCredsResult
		expResult         error
	}{
		{
			name:            "find_creds_error",
			findCredsResult: &findCredsResult{err: errors.New("internal error")},
			expResult:       errors.New("AddNewCreds: can't find dublicates"),
		},
		{
			name:            "creds_already_exist",
			findCredsResult: &findCredsResult{existedCreds: creds.ServiceDAO{Payload: "payload"}},
			expResult:       errors.New("ClientError: creds already exist"),
		},
		{
			name:              "add_creds_error",
			findCredsResult:   &findCredsResult{err: errNoRows},
			addNewCredsResult: &addNewCredsResult{err: errors.New("internal error")},
			expResult:         errors.New("AddNewCreds: can't add creds to db"),
		},
		{
			name:              "success",
			findCredsResult:   &findCredsResult{err: errNoRows},
			addNewCredsResult: &addNewCredsResult{err: nil},
			expResult:         nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository.EXPECT().
				FindCreds(ctx, userID, serviceCreds.Name).
				Return(test.findCredsResult.existedCreds, test.findCredsResult.err).
				Times(1)

			if test.findCredsResult.err != nil {
				isEmpty := mockRepository.EXPECT().IsEmptyRows(test.findCredsResult.err)
				if test.findCredsResult.err == errNoRows {
					isEmpty.Return(true).Times(1)
				} else {
					isEmpty.Return(false).Times(1)
				}
			}

			if test.addNewCredsResult != nil {
				mockRepository.EXPECT().
					AddNewCreds(ctx, gomock.AssignableToTypeOf(creds.ServiceDAO{})).
					Return(test.addNewCredsResult.err).
					Times(1)
			}

			err := usecase.AddNewCreds(ctx, userID, serviceCreds)

			if got, want := err, test.expResult; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestGetCreds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := mock_usecases.NewMockrepository(ctrl)
	ciphers := generateTestCiphers()
	usecase := New(mockRepository, ciphers)

	ctx := context.Background()
	userID := uuid.New()
	serviceName := "test_service_name"
	errNoRows := errors.New("no rows")

	serviceCredsDAO := creds.ServiceDAO{
		ID:      uuid.New(),
		Owner:   userID,
		Name:    serviceName,
		Key:     0,
		Payload: "4ee9ffa37549fbddda83823e5178e0a75065b14add39608f707ba51d1b07056a11a5359b658d19908bd6221d9bca1130",
	}

	serviceCreds := creds.Service{
		Name:     serviceName,
		Login:    "service_login",
		Password: "service_password",
	}

	type findCredsResult struct {
		existedCreds creds.ServiceDAO
		err          error
	}

	type expResult struct {
		serviceCreds creds.Service
		err          error
	}

	tests := []struct {
		name            string
		findCredsResult *findCredsResult
		expResult       expResult
	}{
		{
			name:            "find_creds_error",
			findCredsResult: &findCredsResult{err: errors.New("internal error")},
			expResult:       expResult{err: errors.New("GetCreds: can't get creds from db")},
		},
		{
			name:            "creds_not_found",
			findCredsResult: &findCredsResult{err: errNoRows},
			expResult:       expResult{err: errors.New("ClientError: creds not found")},
		},
		{
			name:            "success",
			findCredsResult: &findCredsResult{existedCreds: serviceCredsDAO},
			expResult:       expResult{serviceCreds: serviceCreds},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository.EXPECT().
				FindCreds(ctx, userID, serviceName).
				Return(test.findCredsResult.existedCreds, test.findCredsResult.err).
				Times(1)

			if test.findCredsResult.err != nil {
				isEmpty := mockRepository.EXPECT().IsEmptyRows(test.findCredsResult.err)
				if test.findCredsResult.err == errNoRows {
					isEmpty.Return(true).Times(1)
				} else {
					isEmpty.Return(false).Times(1)
				}
			}

			actualCreds, actualErr := usecase.GetCreds(ctx, userID, serviceName)

			if got, want := actualErr, test.expResult.err; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if got, want := actualCreds, test.expResult.serviceCreds; got != want {
				t.Fatalf("Wrong! Unexpected creds!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestUpdateCreds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := mock_usecases.NewMockrepository(ctrl)
	ciphers := generateTestCiphers()
	usecase := New(mockRepository, ciphers)

	userID := uuid.New()
	oldServiceName := "old_service_name"
	updatedCreds := creds.Service{
		Name:     "service_name",
		Login:    "service_login",
		Password: "service_password",
	}
	errNoRows := errors.New("no rows")
	ctx := context.Background()

	tests := []struct {
		name              string
		updateCredsResult error
		expResult         error
	}{
		{
			name:              "creds_not_found",
			updateCredsResult: errNoRows,
			expResult:         errors.New("ClientError: creds not found"),
		},
		{
			name:              "update_creds_error",
			updateCredsResult: errors.New("internal error"),
			expResult:         errors.New("UpdateCreds: can't update creds"),
		},
		{
			name:              "success",
			updateCredsResult: nil,
			expResult:         nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepository.EXPECT().
				UpdateCreds(ctx, oldServiceName, gomock.AssignableToTypeOf(creds.ServiceDAO{})).
				Return(test.updateCredsResult).
				Times(1)

			if test.updateCredsResult != nil {
				isEmpty := mockRepository.EXPECT().IsEmptyRows(test.updateCredsResult)
				if test.updateCredsResult == errNoRows {
					isEmpty.Return(true).Times(1)
				} else {
					isEmpty.Return(false).Times(1)
				}
			}

			actualErr := usecase.UpdateCreds(ctx, userID, oldServiceName, updatedCreds)

			if got, want := actualErr, test.expResult; !errors.Is(got, want) {
				t.Fatalf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}
