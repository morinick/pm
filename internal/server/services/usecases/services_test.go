package usecases

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"passman/internal/server/services"
	mock_usecases "passman/internal/server/services/usecases/mock"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestAddService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	saveDir := t.TempDir()
	mockRepo := mock_usecases.NewMockrepository(ctrl)
	serviceUsecase := New(mockRepo, saveDir)
	ctx := context.Background()
	errEmptyRows := errors.New("empty rows")

	type getServiceResult struct {
		service services.Service
		err     error
	}

	type addServiceResult struct {
		err error
	}

	tests := []struct {
		name             string
		getServiceResult *getServiceResult
		addServiceResult *addServiceResult
		expResult        error
	}{
		{
			name: "failed_checking_dublicates",
			getServiceResult: &getServiceResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("AddService: failed checking services dublicates"),
		},
		{
			name: "dublicates_exist",
			getServiceResult: &getServiceResult{
				service: services.Service{Logo: "path/to/logo.svg"},
			},
			expResult: errors.New("ClientError: service with this name already exist"),
		},
		{
			name: "failed_inserting_to_db",
			getServiceResult: &getServiceResult{
				err: errEmptyRows,
			},
			addServiceResult: &addServiceResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("AddService: failed inserting new service"),
		},
		{
			name: "success",
			getServiceResult: &getServiceResult{
				err: errEmptyRows,
			},
			addServiceResult: &addServiceResult{},
			expResult:        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serviceName := "service" + test.name
			readerBody := "some logo"
			r := strings.NewReader(readerBody)

			mockRepo.EXPECT().
				GetService(ctx, serviceName).
				Return(test.getServiceResult.service, test.getServiceResult.err).
				Times(1)

			if test.getServiceResult.err != nil {
				ier := mockRepo.EXPECT().IsEmptyRows(test.getServiceResult.err)
				if test.getServiceResult.err == errEmptyRows {
					ier.Return(true).Times(1)
				} else {
					ier.Return(false).Times(1)
				}
			}

			if test.addServiceResult != nil {
				mockRepo.EXPECT().
					AddService(ctx, gomock.AssignableToTypeOf(services.Service{})).
					Return(test.addServiceResult.err).
					Times(1)
			}

			actErr := serviceUsecase.AddService(ctx, serviceName, r)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if test.name == "success" {
				fileCheck, err := os.ReadFile(filepath.Join(saveDir, serviceName))
				if err != nil {
					t.Fatalf("Wrong! Unexpected error while reading saved file!\n\tError: %v", err)
				}

				if string(fileCheck) != readerBody {
					t.Fatalf("Wrong! Mismatch logo!\n\tExpected: %s\n\tActual: %s", readerBody, string(fileCheck))
				}
			}
		})
	}
}

func TestUpdateService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	saveDir := t.TempDir()
	mockRepo := mock_usecases.NewMockrepository(ctrl)
	serviceUsecase := New(mockRepo, saveDir)
	ctx := context.Background()
	errEmptyRows := errors.New("empty rows")

	type getServiceResult struct {
		service services.Service
		err     error
	}

	type updateServiceResult struct {
		err error
	}

	tests := []struct {
		name                string
		getServiceResult    *getServiceResult
		updateServiceResult *updateServiceResult
		expResult           error
	}{
		{
			name: "failed_checking_dublicates",
			getServiceResult: &getServiceResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateService: failed checking services dublicates"),
		},
		{
			name: "dublicates_exist",
			getServiceResult: &getServiceResult{
				service: services.Service{Logo: "path/to/logo"},
			},
			expResult: errors.New("ClientError: service with this name already exist"),
		},
		{
			name: "failed_updating_db",
			getServiceResult: &getServiceResult{
				err: errEmptyRows,
			},
			updateServiceResult: &updateServiceResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateService: failed updating service in database"),
		},
		{
			name: "success_with_changing_name",
			getServiceResult: &getServiceResult{
				err: errEmptyRows,
			},
			updateServiceResult: &updateServiceResult{},
			expResult:           nil,
		},
		{
			name: "success_without_changing_name",
			getServiceResult: &getServiceResult{
				err: errEmptyRows,
			},
			updateServiceResult: &updateServiceResult{},
			expResult:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldServiceName := "oldName" + test.name
			newServiceName := "newName" + test.name
			if test.name == "success_without_changing_name" {
				newServiceName = oldServiceName
			}

			testFile, err := os.Create(filepath.Join(saveDir, oldServiceName))
			if err != nil {
				t.Fatalf("Unexpected error with creation logo test file: %v", err)
			}
			testFile.Close()
			logoFileBody := "some logo"
			logoFile := strings.NewReader(logoFileBody)

			if test.name != "success_without_changing_name" {
				mockRepo.EXPECT().
					GetService(ctx, newServiceName).
					Return(test.getServiceResult.service, test.getServiceResult.err).
					Times(1)

				if test.getServiceResult.err != nil {
					ier := mockRepo.EXPECT().IsEmptyRows(test.getServiceResult.err)
					if test.getServiceResult.err == errEmptyRows {
						ier.Return(true).Times(1)
					} else {
						ier.Return(false).Times(1)
					}
				}

				if test.updateServiceResult != nil {
					mockRepo.EXPECT().
						UpdateService(ctx, oldServiceName, gomock.AssignableToTypeOf(services.ServiceDTO{})).
						Return(test.updateServiceResult.err).
						Times(1)
				}
			}

			actErr := serviceUsecase.UpdateService(ctx, oldServiceName, newServiceName, logoFile)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if strings.HasPrefix(test.name, "success") {
				fileCheck, err := os.ReadFile(filepath.Join(saveDir, newServiceName))
				if err != nil {
					t.Fatalf("Wrong! Unexpected error while reading saved file!\n\tError: %v", err)
				}

				if string(fileCheck) != logoFileBody {
					t.Fatalf("Wrong! Mismatch logo!\n\tExpected: %s\n\tActual: %s", logoFileBody, string(fileCheck))
				}
			}
		})
	}
}

func TestRemoveService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	saveDir := t.TempDir()
	mockRepo := mock_usecases.NewMockrepository(ctrl)
	serviceUsecase := New(mockRepo, saveDir)
	ctx := context.Background()
	userID := uuid.New()
	errNoRows := errors.New("no rows")

	type checkExistingRecordResult struct {
		credID uuid.UUID
		err    error
	}

	type removeServiceResult struct {
		err error
	}

	tests := []struct {
		name                      string
		checkExistingRecordResult *checkExistingRecordResult
		removeServiceResult       *removeServiceResult
		expResult                 error
	}{
		{
			name:                      "failed_checking_records",
			checkExistingRecordResult: &checkExistingRecordResult{err: errors.New("internal error")},
			expResult:                 errors.New("RemoveService: failed checking another records"),
		},
		{
			name:                      "another_user_has_record",
			checkExistingRecordResult: &checkExistingRecordResult{credID: uuid.New()},
			expResult:                 errors.New("ClientError: another users has records in this service"),
		},
		{
			name:                      "failed_removing_from_db",
			checkExistingRecordResult: &checkExistingRecordResult{err: errNoRows},
			removeServiceResult:       &removeServiceResult{err: errors.New("internal error")},
			expResult:                 errors.New("RemoveService: failed removing service from database"),
		},
		{
			name:                      "success",
			checkExistingRecordResult: &checkExistingRecordResult{err: errNoRows},
			removeServiceResult:       &removeServiceResult{err: nil},
			expResult:                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serviceName := "service" + test.name
			testFile, _ := os.Create(filepath.Join(saveDir, serviceName))
			testFile.Close()

			mockRepo.EXPECT().
				CheckExistingRecord(ctx, userID, serviceName).
				Return(
					test.checkExistingRecordResult.credID,
					test.checkExistingRecordResult.err,
				).
				Times(1)

			if test.checkExistingRecordResult.err != nil {
				ier := mockRepo.EXPECT().IsEmptyRows(test.checkExistingRecordResult.err)
				if test.checkExistingRecordResult.err == errNoRows {
					ier.Return(true).Times(1)
				} else {
					ier.Return(false).Times(1)
				}
			}

			if test.removeServiceResult != nil {
				mockRepo.EXPECT().
					RemoveService(ctx, serviceName).
					Return(test.removeServiceResult.err).
					Times(1)
			}

			actErr := serviceUsecase.RemoveService(ctx, userID, serviceName)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual:%v", want, got)
			}

			if test.name == "success" {
				dirs, _ := os.ReadDir(saveDir)
				for _, checkFile := range dirs {
					if checkFile.Name() == serviceName {
						t.Errorf("Wrong! Logo file has not been deleted!")
					}
				}
			}
		})
	}
}
