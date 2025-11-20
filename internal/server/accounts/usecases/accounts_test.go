package usecases

import (
	"context"
	"errors"
	"testing"

	"passman/internal/server/accounts"
	mock_usecases "passman/internal/server/accounts/usecases/mock"
	"passman/pkg/cipher"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func generateTestCiphers() []cipher.AESCipher {
	hexKey := "5f1e40c065ef8e1c99342e8ca567d12f7825fedf25f10a7636effc9f766e7013"
	ciph, _ := cipher.New(hexKey)
	return []cipher.AESCipher{*ciph}
}

func compareDTOs(s1, s2 []accounts.AccountDTO) bool {
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

func TestAddAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	accountsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()
	errNoRows := errors.New("no rows")
	serviceID := uuid.New()
	userID := uuid.New()

	type getServiceIDResult struct {
		serviceID uuid.UUID
		err       error
	}

	type getAccountIDResult struct {
		dublicateID uuid.UUID
		err         error
	}

	type addAccountResult struct {
		err error
	}

	tests := []struct {
		name               string
		input              accounts.AccountDTO
		getServiceIDResult *getServiceIDResult
		getAccountIDResult *getAccountIDResult
		addAccountResult   *addAccountResult
		expResult          error
	}{
		{
			name: "failed_getting_service_id",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{ServiceName: "ServiceName"},
			},
			getServiceIDResult: &getServiceIDResult{err: errors.New("internal error")},
			expResult:          errors.New("AddAccount: failed getting service id"),
		},
		{
			name: "invalid_service_name",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{ServiceName: "ServiceName"},
			},
			getServiceIDResult: &getServiceIDResult{err: errNoRows},
			expResult:          errors.New("ClientError: invalid service name"),
		},
		{
			name: "failed_checking_dublicates",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name: "accName",
			},
			getServiceIDResult: &getServiceIDResult{serviceID: serviceID},
			getAccountIDResult: &getAccountIDResult{err: errors.New("internal error")},
			expResult:          errors.New("AddAccount: failed checking dublicates"),
		},
		{
			name: "account_already_exist",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name: "accName",
			},
			getServiceIDResult: &getServiceIDResult{serviceID: serviceID},
			getAccountIDResult: &getAccountIDResult{dublicateID: uuid.New()},
			expResult:          errors.New("ClientError: account with this name already exist"),
		},
		{
			name: "failed_adding_account",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "accName",
				Login:    "login",
				Password: "password",
			},
			getServiceIDResult: &getServiceIDResult{serviceID: serviceID},
			getAccountIDResult: &getAccountIDResult{dublicateID: uuid.Nil},
			addAccountResult:   &addAccountResult{err: errors.New("internal error")},
			expResult:          errors.New("AddAccount: failed adding account"),
		},
		{
			name: "success",
			input: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
					UserID:      userID,
				},
				Name:     "credRecordName",
				Login:    "login",
				Password: "password",
			},
			getServiceIDResult: &getServiceIDResult{serviceID: serviceID},
			getAccountIDResult: &getAccountIDResult{dublicateID: uuid.Nil},
			addAccountResult:   &addAccountResult{err: nil},
			expResult:          nil,
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

			if test.getAccountIDResult != nil {
				mockRepo.EXPECT().
					GetAccountID(
						ctx,
						test.input.UserID,
						gomock.AssignableToTypeOf(uuid.UUID{}),
						test.input.Name,
					).
					Return(
						test.getAccountIDResult.dublicateID,
						test.getAccountIDResult.err,
					).
					Times(1)

				if test.getAccountIDResult.err != nil {
					ier := mockRepo.EXPECT().IsEmptyRows(test.getAccountIDResult.err)
					if test.getAccountIDResult.err == errNoRows {
						ier.Return(true).Times(1)
					} else {
						ier.Return(false).Times(1)
					}
				}
			}

			if test.addAccountResult != nil {
				mockRepo.EXPECT().
					AddAccount(ctx, gomock.AssignableToTypeOf(accounts.Account{})).
					Return(test.addAccountResult.err).
					Times(1)
			}

			actErr := accountsUsecase.AddAccount(ctx, test.input)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestGetAccountsInService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	accountsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()

	inputParams := accounts.QueryParams{
		UserID:      uuid.New(),
		ServiceName: "some_service",
	}

	incorrectAccounts := []accounts.Account{
		{
			UserID:    inputParams.UserID,
			ServiceID: uuid.New(),
			Name:      "name",
			Secret:    0,
			Payload:   "incorrect_payload",
		},
	}

	correctAccounts := []accounts.Account{
		{
			UserID:    inputParams.UserID,
			ServiceID: uuid.New(),
			Name:      "acc_name",
			Secret:    0,
			Payload:   "c68acc479af6a2caa531d56def51a3caf97304f691440c40d8f0297570c20f2a",
		},
	}

	type getAccountsResult struct {
		records []accounts.Account
		err     error
	}

	type expResult struct {
		dtos []accounts.AccountDTO
		err  error
	}

	tests := []struct {
		name              string
		getAccountsResult getAccountsResult
		expResult         expResult
	}{
		{
			name: "failed_getting_records",
			getAccountsResult: getAccountsResult{
				err: errors.New("internal error"),
			},
			expResult: expResult{
				err: errors.New("GetAccountsInService: failed getting accounts"),
			},
		},
		{
			name: "failed_decrypting",
			getAccountsResult: getAccountsResult{
				records: incorrectAccounts,
			},
			expResult: expResult{
				err: errors.New("GetAccountsInService: failed decrypting account"),
			},
		},
		{
			name: "success",
			getAccountsResult: getAccountsResult{
				records: correctAccounts,
			},
			expResult: expResult{
				dtos: []accounts.AccountDTO{
					{
						QueryParams: accounts.QueryParams{
							UserID: inputParams.UserID,
						},
						Name:     "acc_name",
						Login:    "acc_login",
						Password: "acc_password",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetUserAccountsInService(ctx, inputParams).
				Return(test.getAccountsResult.records, test.getAccountsResult.err).
				Times(1)

			actDTOs, actErr := accountsUsecase.GetAccountsInService(ctx, inputParams)

			if got, want := actDTOs, test.expResult.dtos; !compareDTOs(got, want) {
				t.Errorf("Wrong! Mismatch account dtos!\n\tExpected: %v\n\tActual: %v", want, got)
			}

			if got, want := actErr, test.expResult.err; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecases.NewMockrepository(ctrl)
	testCiphers := generateTestCiphers()
	accountsUsecase := New(mockRepo, testCiphers)

	ctx := context.Background()
	errNoRows := errors.New("no rows")
	serviceID := uuid.New()
	userID := uuid.New()

	type getServiceIDResult struct {
		serviceID uuid.UUID
		err       error
	}

	type getAccountIDResult struct {
		accountID uuid.UUID
		err       error
	}

	type updateAccountResult struct {
		err error
	}

	tests := []struct {
		name                string
		oldAccountName      string
		updatedAccount      accounts.AccountDTO
		getServiceIDResult  *getServiceIDResult
		getAccountIDResult  *getAccountIDResult
		updateAccountResult *updateAccountResult
		expResult           error
	}{
		{
			name:           "failed_getting_service_id",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
				},
			},
			getServiceIDResult: &getServiceIDResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateAccount: failed getting service id"),
		},
		{
			name:           "invalid_service_name",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
					ServiceName: "ServiceName",
				},
			},
			getServiceIDResult: &getServiceIDResult{
				err: errNoRows,
			},
			expResult: errors.New("ClientError: invalid service name"),
		},
		{
			name:           "failed_checking_old_name",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
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
			getAccountIDResult: &getAccountIDResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateAccount: failed checking old account name"),
		},
		{
			name:           "invalid_old_name",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
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
			getAccountIDResult: &getAccountIDResult{
				err: errNoRows,
			},
			expResult: errors.New("ClientError: invalid old account name"),
		},
		{
			name:           "failed_updating_cred_record",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
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
			getAccountIDResult: &getAccountIDResult{
				err: nil,
			},
			updateAccountResult: &updateAccountResult{
				err: errors.New("internal error"),
			},
			expResult: errors.New("UpdateAccount: failed updating account"),
		},
		{
			name:           "success",
			oldAccountName: "oldName",
			updatedAccount: accounts.AccountDTO{
				QueryParams: accounts.QueryParams{
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
			getAccountIDResult: &getAccountIDResult{
				err: nil,
			},
			updateAccountResult: &updateAccountResult{
				err: nil,
			},
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().
				GetServiceID(ctx, test.updatedAccount.ServiceName).
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

			if test.getAccountIDResult != nil {
				mockRepo.EXPECT().
					GetAccountID(
						ctx,
						test.updatedAccount.UserID,
						test.getServiceIDResult.serviceID,
						test.oldAccountName,
					).
					Return(test.getAccountIDResult.accountID, test.getAccountIDResult.err).
					Times(1)

				if test.getAccountIDResult.err != nil {
					iet := mockRepo.EXPECT().IsEmptyRows(test.getAccountIDResult.err)
					if test.getAccountIDResult.err == errNoRows {
						iet.Return(true).Times(1)
					} else {
						iet.Return(false).Times(1)
					}
				}
			}

			if test.updateAccountResult != nil {
				mockRepo.EXPECT().
					UpdateAccount(
						ctx,
						test.oldAccountName,
						gomock.AssignableToTypeOf(accounts.Account{}),
					).
					Return(test.updateAccountResult.err).
					Times(1)
			}

			actErr := accountsUsecase.UpdateAccount(ctx, test.oldAccountName, test.updatedAccount)

			if got, want := actErr, test.expResult; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActual: %v", want, got)
			}
		})
	}
}
