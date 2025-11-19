package usecases

import (
	"context"
	"errors"
	"testing"

	"passman/internal/server/creds"
	mock_usecases "passman/internal/server/creds/usecases/mock"
	"passman/pkg/cipher"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func generateTestCiphers() []cipher.AESCipher {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	ciph, _ := cipher.New(hexKey)
	return []cipher.AESCipher{*ciph}
}

func compareRecordTransfers(s1, s2 []creds.CredsRecordTransfer) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i].Name != s2[i].Name {
			return false
		}
		if s1[i].Login != s2[i].Login {
			return false
		}
		if s1[i].Password != s2[i].Password {
			return false
		}
		if s1[i].UserID != s2[i].UserID {
			return false
		}
	}
	return true
}

func TestAddCredsRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	credsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()
	errNoRows := errors.New("no rows")
	serviceID := uuid.New()
	userID := uuid.New()

	type getServiceIDResult struct {
		serviceID uuid.UUID
		err       error
	}

	type getCredsRecordIDResult struct {
		dublicateID uuid.UUID
		err         error
	}

	type addServiceCredsResult struct {
		err error
	}

	tests := []struct {
		name                   string
		input                  creds.CredsRecordTransfer
		getServiceIDResult     *getServiceIDResult
		getCredsRecordIDResult *getCredsRecordIDResult
		addServiceCredsResult  *addServiceCredsResult
		expResult              error
	}{
		{
			name: "failed_getting_service_id",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{ServiceName: "ServiceName"},
			},
			getServiceIDResult: &getServiceIDResult{err: errors.New("internal error")},
			expResult:          errors.New("AddCredsRecord: failed getting service id"),
		},
		{
			name: "invalid_service_name",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{ServiceName: "ServiceName"},
			},
			getServiceIDResult: &getServiceIDResult{err: errNoRows},
			expResult:          errors.New("ClientError: invalid service name"),
		},
		{
			name: "failed_checking_dublicates",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name: "credRecordName",
			},
			getServiceIDResult:     &getServiceIDResult{serviceID: serviceID},
			getCredsRecordIDResult: &getCredsRecordIDResult{err: errors.New("internal error")},
			expResult:              errors.New("AddCredsRecord: failed checking dublicates"),
		},
		{
			name: "record_already_exist",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name: "credRecordName",
			},
			getServiceIDResult:     &getServiceIDResult{serviceID: serviceID},
			getCredsRecordIDResult: &getCredsRecordIDResult{dublicateID: uuid.New()},
			expResult:              errors.New("ClientError: creds with this name already exist"),
		},
		{
			name: "failed_adding_record",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "credRecordName",
				Login:    "login",
				Password: "password",
			},
			getServiceIDResult:     &getServiceIDResult{serviceID: serviceID},
			getCredsRecordIDResult: &getCredsRecordIDResult{dublicateID: uuid.Nil},
			addServiceCredsResult:  &addServiceCredsResult{err: errors.New("internal error")},
			expResult:              errors.New("AddCredsRecord: failed adding creds record"),
		},
		{
			name: "success",
			input: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "credRecordName",
				Login:    "login",
				Password: "password",
			},
			getServiceIDResult:     &getServiceIDResult{serviceID: serviceID},
			getCredsRecordIDResult: &getCredsRecordIDResult{dublicateID: uuid.Nil},
			addServiceCredsResult:  &addServiceCredsResult{err: nil},
			expResult:              nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetServiceID(ctx, test.input.ServiceName).
				Return(test.getServiceIDResult.serviceID, test.getServiceIDResult.err).
				Times(1)

			if test.getServiceIDResult.err != nil {
				ier := mockRepo.EXPECT().IsEmptyRows(test.getServiceIDResult.err)
				if test.getServiceIDResult.err == errNoRows {
					ier.Return(true).Times(1)
				} else {
					ier.Return(false).Times(1)
				}
			}

			if test.getCredsRecordIDResult != nil {
				mockRepo.EXPECT().
					GetCredsRecordID(
						ctx,
						test.input.UserID,
						gomock.AssignableToTypeOf(uuid.UUID{}),
						test.input.Name,
					).
					Return(
						test.getCredsRecordIDResult.dublicateID,
						test.getCredsRecordIDResult.err,
					).
					Times(1)

				if test.getCredsRecordIDResult.err != nil {
					ier := mockRepo.EXPECT().IsEmptyRows(test.getCredsRecordIDResult.err)
					if test.getCredsRecordIDResult.err == errNoRows {
						ier.Return(true).Times(1)
					} else {
						ier.Return(false).Times(1)
					}
				}
			}

			if test.addServiceCredsResult != nil {
				mockRepo.EXPECT().
					AddCredsRecord(ctx, gomock.AssignableToTypeOf(creds.CredsRecord{})).
					Return(test.addServiceCredsResult.err).
					Times(1)
			}

			actErr := credsUsecase.AddCredsRecord(ctx, test.input)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestGetCredsRecordsInService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	credsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()

	inputParams := creds.QueryParams{
		UserID:      uuid.New(),
		ServiceName: "some_service",
	}

	incorrectRecords := []creds.CredsRecord{
		{
			UserID:    inputParams.UserID,
			ServiceID: uuid.New(),
			Name:      "name",
			Secret:    0,
			Payload:   "incorrect_payload",
		},
	}

	correctRecords := []creds.CredsRecord{
		{
			UserID:    inputParams.UserID,
			ServiceID: uuid.New(),
			Name:      "cred_name",
			Secret:    0,
			Payload:   "b5a0bfeee22432ea139f70fa13a164ab9b1019f0e2554d2efd75acef50910050",
		},
	}

	type getUserCredsResult struct {
		records []creds.CredsRecord
		err     error
	}

	type expResult struct {
		crts []creds.CredsRecordTransfer
		err  error
	}

	tests := []struct {
		name               string
		getUserCredsResult getUserCredsResult
		expResult          expResult
	}{
		{
			name: "failed_getting_records",
			getUserCredsResult: getUserCredsResult{
				err: errors.New("internal error"),
			},
			expResult: expResult{
				err: errors.New("GetCredsRecordsInService: failed getting records"),
			},
		},
		{
			name: "failed_decrypting",
			getUserCredsResult: getUserCredsResult{
				records: incorrectRecords,
			},
			expResult: expResult{
				err: errors.New("GetCredsRecordsInService: failed decrypting record"),
			},
		},
		{
			name: "success",
			getUserCredsResult: getUserCredsResult{
				records: correctRecords,
			},
			expResult: expResult{
				crts: []creds.CredsRecordTransfer{
					{
						QueryParams: creds.QueryParams{
							UserID: inputParams.UserID,
						},
						Name:     "cred_name",
						Login:    "cred_login",
						Password: "cred_password",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUserCredsInService(ctx, inputParams).
				Return(test.getUserCredsResult.records, test.getUserCredsResult.err).
				Times(1)

			actCrts, actErr := credsUsecase.GetCredsRecordsInService(ctx, inputParams)

			if got, want := actCrts, test.expResult.crts; !compareRecordTransfers(got, want) {
				t.Errorf("Wrong! Mismatch record transfers!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if got, want := actErr, test.expResult.err; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestUpdateCredsRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	credsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()
	errNoRows := errors.New("no rows")
	serviceID := uuid.New()
	userID := uuid.New()

	type getServiceIDResult struct {
		serviceID uuid.UUID
		err       error
	}

	type getCredsRecordIDResult struct {
		credID uuid.UUID
		err    error
	}

	type updateServiceCredsResult struct {
		err error
	}

	tests := []struct {
		name                     string
		oldCredRecordName        string
		updatedCredRecord        creds.CredsRecordTransfer
		getServiceIDResult       *getServiceIDResult
		getCredsRecordIDResult   *getCredsRecordIDResult
		updateServiceCredsResult *updateServiceCredsResult
		expResult                error
	}{
		{
			name:              "failed_getting_service_id",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
				},
			},
			getServiceIDResult: &getServiceIDResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateCredsRecord: failed getting service id"),
		},
		{
			name:              "invalid_service_name",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
				},
			},
			getServiceIDResult: &getServiceIDResult{
				err: errNoRows,
			},
			expResult: errors.New("ClientError: invalid service name"),
		},
		{
			name:              "failed_checking_old_name",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "SomeName",
				Login:    "SomeLogin",
				Password: "SomePassword",
			},
			getServiceIDResult: &getServiceIDResult{
				serviceID: serviceID,
			},
			getCredsRecordIDResult: &getCredsRecordIDResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateCredsRecord: failed checking old cred record name"),
		},
		{
			name:              "invalid_old_name",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "SomeName",
				Login:    "SomeLogin",
				Password: "SomePassword",
			},
			getServiceIDResult: &getServiceIDResult{
				serviceID: serviceID,
			},
			getCredsRecordIDResult: &getCredsRecordIDResult{
				err: errNoRows,
			},
			expResult: errors.New("ClientError: invalid old cred name"),
		},
		{
			name:              "failed_updating_cred_record",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "SomeName",
				Login:    "SomeLogin",
				Password: "SomePassword",
			},
			getServiceIDResult: &getServiceIDResult{
				serviceID: serviceID,
			},
			getCredsRecordIDResult: &getCredsRecordIDResult{
				err: nil,
			},
			updateServiceCredsResult: &updateServiceCredsResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateCredsRecord: failed updating cred record"),
		},
		{
			name:              "success",
			oldCredRecordName: "oldName",
			updatedCredRecord: creds.CredsRecordTransfer{
				QueryParams: creds.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "SomeName",
				Login:    "SomeLogin",
				Password: "SomePassword",
			},
			getServiceIDResult: &getServiceIDResult{
				serviceID: serviceID,
			},
			getCredsRecordIDResult: &getCredsRecordIDResult{
				err: nil,
			},
			updateServiceCredsResult: &updateServiceCredsResult{
				err: nil,
			},
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetServiceID(ctx, test.updatedCredRecord.ServiceName).
				Return(test.getServiceIDResult.serviceID, test.getServiceIDResult.err).
				Times(1)

			if test.getServiceIDResult.err != nil {
				iet := mockRepo.EXPECT().IsEmptyRows(test.getServiceIDResult.err)
				if test.getServiceIDResult.err == errNoRows {
					iet.Return(true).Times(1)
				} else {
					iet.Return(false).Times(1)
				}
			}

			if test.getCredsRecordIDResult != nil {
				mockRepo.EXPECT().
					GetCredsRecordID(
						ctx,
						test.updatedCredRecord.UserID,
						test.getServiceIDResult.serviceID,
						test.oldCredRecordName,
					).
					Return(test.getCredsRecordIDResult.credID, test.getCredsRecordIDResult.err).
					Times(1)

				if test.getCredsRecordIDResult.err != nil {
					iet := mockRepo.EXPECT().IsEmptyRows(test.getCredsRecordIDResult.err)
					if test.getCredsRecordIDResult.err == errNoRows {
						iet.Return(true).Times(1)
					} else {
						iet.Return(false).Times(1)
					}
				}
			}

			if test.updateServiceCredsResult != nil {
				mockRepo.EXPECT().
					UpdateCredsRecord(
						ctx,
						test.oldCredRecordName,
						gomock.AssignableToTypeOf(creds.CredsRecord{}),
					).
					Return(test.updateServiceCredsResult.err).
					Times(1)
			}

			actErr := credsUsecase.UpdateCredsRecord(ctx, test.oldCredRecordName, test.updatedCredRecord)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}
