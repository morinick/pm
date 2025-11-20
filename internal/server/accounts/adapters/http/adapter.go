package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"passman/internal/server/accounts"
	"passman/internal/server/infra"

	"github.com/go-chi/chi/v5"
	vldtr "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Adapter struct {
	log *slog.Logger
	cu  accountsUsecases
	sm  sessionManager
	v   *validator
}

func NewRouter(cu accountsUsecases, sm sessionManager, v *vldtr.Validate) chi.Router {
	a := &Adapter{
		log: slog.Default(),
		cu:  cu,
		sm:  sm,
		v:   newValidator(v),
	}

	router := chi.NewRouter()

	router.Use(infra.AuthMiddleware(sm))

	router.Post("/{serviceName}", a.AddAccount)
	router.Get("/{serviceName}", a.GetAccountsInService)
	router.Put("/{serviceName}", a.UpdateAccount)
	router.Delete("/{serviceName}/{accountName}", a.RemoveAccount)
	router.Delete("/{serviceName}", a.RemoveAllAccountsInService)

	return router
}

func (a *Adapter) AddAccount(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	body := struct {
		Name     string `json:"name"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "AddAccount: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateAccount(body.Name, body.Login, body.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	transfer := accounts.AccountDTO{
		QueryParams: accounts.QueryParams{
			ServiceName: serviceName,
			UserID:      userID,
		},
		Name:     body.Name,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.cu.AddAccount(r.Context(), transfer); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "AddAccount", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) GetAccountsInService(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	accounts, err := a.cu.GetAccountsInService(r.Context(), accounts.QueryParams{UserID: userID, ServiceName: serviceName})
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "GetAccountsInService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	type responseType struct {
		Name     string `json:"name"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	res := make([]responseType, 0, len(accounts))
	for _, acc := range accounts {
		res = append(res, responseType{Name: acc.Name, Login: acc.Login, Password: acc.Password})
	}

	infra.ResponseJSON(w, res, http.StatusOK)
}

func (a *Adapter) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	body := struct {
		OldName  string `json:"old_name"`
		NewName  string `json:"new_name"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "UpdateAccount: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateAccount(body.NewName, body.Login, body.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	dto := accounts.AccountDTO{
		QueryParams: accounts.QueryParams{
			ServiceName: serviceName,
			UserID:      userID,
		},
		Name:     body.NewName,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.cu.UpdateAccount(r.Context(), body.OldName, dto); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdateAccount", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveAccount(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	accountName := chi.URLParam(r, "accountName")
	if err := a.v.ValidateName(accountName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.RemoveAccount(r.Context(), accountName, accounts.QueryParams{ServiceName: serviceName, UserID: userID}); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "RemoveAccount", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveAllAccountsInService(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.RemoveAllAccountsInService(r.Context(), accounts.QueryParams{ServiceName: serviceName, UserID: userID}); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "RemoveAllAccountsInService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) ParseUsecaseError(ctx context.Context, component string, usecaseError error) (int, string) {
	code, msg, err := a.cu.ParseMyError(usecaseError)
	if code == 0 {
		a.log.ErrorContext(ctx, fmt.Sprintf("%s: incorrect type of usecase error", component), slog.Any("error", err))
		return http.StatusInternalServerError, "internal error"
	}

	if code >= 500 {
		a.log.ErrorContext(ctx, msg, slog.Any("error", err))
		return code, "internal error"
	}

	a.log.WarnContext(ctx, msg)
	return code, strings.Split(msg, ": ")[1]
}
