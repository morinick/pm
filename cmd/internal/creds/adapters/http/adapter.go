package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"passman/cmd/internal/creds"
	"passman/cmd/internal/infra"
	"strings"

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
	router.Use(
		infra.AuthMiddleware(sm),
	)
	router.Post("/create", a.AddNewCreds)
	router.Get("/get/{name}", a.GetCreds)
	router.Get("/list", a.GetCredsList)
	router.Put("/update/{name}", a.UpdateCreds)
	router.Delete("/delete/{name}", a.RemoveCreds)

	return router
}

func (a *Adapter) AddNewCreds(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	body := struct {
		Name     string `json:"service_name"`
		Login    string `json:"service_login"`
		Password string `json:"service_password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "AddNewCreds: can't parse body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	serviceCreds := creds.Service{
		Name:     body.Name,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.v.ValidateCreds(serviceCreds); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.AddNewCreds(r.Context(), userID, serviceCreds); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "AddNewCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) GetCreds(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "name")
	if err := a.v.ValidateServiceName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	serviceCreds, err := a.cu.GetCreds(r.Context(), userID, serviceName)
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "GetCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	infra.ResponseJSON(w, struct {
		Name     string `json:"service_name"`
		Login    string `json:"service_login"`
		Password string `json:"service_password"`
	}{Name: serviceCreds.Name, Login: serviceCreds.Login, Password: serviceCreds.Password}, http.StatusOK)
}

func (a *Adapter) GetCredsList(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	list, err := a.cu.GetCredsList(r.Context(), userID)
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "GetCredsList", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	infra.ResponseJSON(w, struct {
		List []string `json:"list_of_creds"`
	}{List: list}, http.StatusOK)
}

func (a *Adapter) UpdateCreds(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	oldServiceName := chi.URLParam(r, "name")
	if err := a.v.ValidateServiceName(oldServiceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	body := struct {
		Name     string `json:"service_name"`
		Login    string `json:"service_login"`
		Password string `json:"service_password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "UpdateCreds: can't parse body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	updatedCreds := creds.Service{
		Name:     body.Name,
		Login:    body.Login,
		Password: body.Password,
	}

	if err := a.v.ValidateCreds(updatedCreds); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.UpdateCreds(r.Context(), userID, oldServiceName, updatedCreds); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdateCreds", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveCreds(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "name")
	if err := a.v.ValidateServiceName(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.cu.RemoveCreds(r.Context(), userID, serviceName); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "RemoveCreds", err)
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
