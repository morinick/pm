package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"passman/internal/server/creds"
	"passman/internal/server/infra"

	"github.com/go-chi/chi/v5"
	vldtr "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Adapter struct {
	log *slog.Logger
	cu  credsUsecases
	sm  sessionManager
	v   *validator
}

func NewRouter(cu credsUsecases, sm sessionManager, v *vldtr.Validate) chi.Router {
	a := &Adapter{
		log: slog.Default(),
		cu:  cu,
		sm:  sm,
		v:   newValidator(v),
	}

	router := chi.NewRouter()

	router.Use(infra.AuthMiddleware(sm))

	router.Post("/{serviceName}", a.AddCredsRecord)
	router.Get("/{serviceName}", a.GetCredsRecordsInService)
	router.Put("/{serviceName}", a.UpdateCredsRecord)
	router.Delete("/{serviceName}/{credName}", a.RemoveCredRecord)
	router.Delete("/{serviceName}", a.RemoveAllCredsInService)

	return router
}

func (a *Adapter) AddCredsRecord(w http.ResponseWriter, r *http.Request) {
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
		a.log.ErrorContext(r.Context(), "AddNewServiceCreds: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateCreds(body.Name, body.Login, body.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	transfer := creds.CredsRecordTransfer{
		QueryParams: creds.QueryParams{
			ServiceName: serviceName,
			UserID:      userID,
		},
		Name:     body.Name,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.cu.AddCredsRecord(r.Context(), transfer); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "AddNewServiceCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) GetCredsRecordsInService(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	serviceCreds, err := a.cu.GetCredsRecordsInService(r.Context(), creds.QueryParams{UserID: userID, ServiceName: serviceName})
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "GetServiceCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	type responseType struct {
		Name     string `json:"name"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	res := make([]responseType, 0, len(serviceCreds))
	for _, cred := range serviceCreds {
		res = append(res, responseType{Name: cred.Name, Login: cred.Login, Password: cred.Password})
	}

	infra.ResponseJSON(w, res, http.StatusOK)
}

func (a *Adapter) UpdateCredsRecord(w http.ResponseWriter, r *http.Request) {
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
		a.log.ErrorContext(r.Context(), "UpdateCreds: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateCreds(body.NewName, body.Login, body.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	transfer := creds.CredsRecordTransfer{
		QueryParams: creds.QueryParams{
			ServiceName: serviceName,
			UserID:      userID,
		},
		Name:     body.NewName,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.cu.UpdateCredsRecord(r.Context(), body.OldName, transfer); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdateCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveCredRecord(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	credName := chi.URLParam(r, "credName")
	if err := a.v.ValidateName(credName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.RemoveCredsRecord(r.Context(), credName, creds.QueryParams{ServiceName: serviceName, UserID: userID}); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "RemoveCredRecord", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveAllCredsInService(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.RemoveAllCredsInService(r.Context(), creds.QueryParams{ServiceName: serviceName, UserID: userID}); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "RemoveAllCredsInService", err)
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
