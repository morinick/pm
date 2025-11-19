package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"passman/internal/server/infra"

	"github.com/go-chi/chi/v5"
	vldtr "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Adapter struct {
	log *slog.Logger
	su  serviceUsecase
	sm  sessionManager
	v   *validator
}

func NewRouter(su serviceUsecase, sm sessionManager, v *vldtr.Validate) chi.Router {
	a := &Adapter{
		log: slog.Default(),
		su:  su,
		sm:  sm,
		v:   newValidator(v),
	}

	router := chi.NewRouter()

	router.Use(infra.AuthMiddleware(sm))

	router.Post("/{serviceName}", a.AddService)
	router.Get("/all", a.GetAllServices)
	router.Get("/my", a.GetAllUserServices)
	router.Put("/{oldServiceName}/{newServiceName}", a.UpdateService)
	router.Delete("/{serviceName}", a.RemoveService)

	return router
}

func (a *Adapter) AddService(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateServiceNames(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	file, err := infra.RecieveFile(
		r,
		infra.RecieveFileOptions{
			FormFileKey:       "logo",
			ValidContentTypes: []string{"image/jpeg", "image/png", "image/svg+xml"},
		})
	if err != nil {
		code, msg := a.parseRecieveFileError(r.Context(), "AddService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}
	defer file.Close()

	if err := a.su.AddService(r.Context(), serviceName, file); err != nil {
		code, msg := a.parseUsecaseError(r.Context(), "AddService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) GetAllServices(w http.ResponseWriter, r *http.Request) {
	services, err := a.su.GetAllServices(r.Context())
	if err != nil {
		code, msg := a.parseUsecaseError(r.Context(), "GetAllServices", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	type responseType struct {
		Name     string `json:"name"`
		LogoPath string `json:"logo_path"`
	}
	res := make([]responseType, 0, len(services))
	for _, serv := range services {
		res = append(res, responseType{Name: serv.Name, LogoPath: serv.Logo})
	}

	infra.ResponseJSON(w, res, http.StatusOK)
}

func (a *Adapter) GetAllUserServices(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	services, err := a.su.GetAllUserServices(r.Context(), userID)
	if err != nil {
		code, msg := a.parseUsecaseError(r.Context(), "GetAllUserServices", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	type responseType struct {
		Name     string `json:"name"`
		LogoPath string `json:"logo_path"`
	}

	res := make([]responseType, 0, len(services))
	for _, serv := range services {
		res = append(res, responseType{Name: serv.Name, LogoPath: serv.Logo})
	}

	infra.ResponseJSON(w, res, http.StatusOK)
}

func (a *Adapter) UpdateService(w http.ResponseWriter, r *http.Request) {
	oldServiceName := chi.URLParam(r, "oldServiceName")
	newServiceName := chi.URLParam(r, "newServiceName")

	if err := a.v.ValidateServiceNames(oldServiceName, newServiceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	file, err := infra.RecieveFile(
		r,
		infra.RecieveFileOptions{
			FormFileKey:       "logo",
			ValidContentTypes: []string{"image/jpeg", "image/png", "image/svg+xml"},
		})
	if err != nil {
		code, msg := a.parseRecieveFileError(r.Context(), "UpdateService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}
	defer file.Close()

	if err := a.su.UpdateService(r.Context(), oldServiceName, newServiceName, file); err != nil {
		code, msg := a.parseUsecaseError(r.Context(), "UpdateService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveService(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.sm.GetString(r.Context(), "user_id"))

	serviceName := chi.URLParam(r, "serviceName")
	if err := a.v.ValidateServiceNames(serviceName); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.su.RemoveService(r.Context(), userID, serviceName); err != nil {
		code, msg := a.parseUsecaseError(r.Context(), "RemoveService", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) parseUsecaseError(ctx context.Context, component string, usecaseError error) (int, string) {
	code, msg, err := a.su.ParseUserError(usecaseError)
	if code == 0 {
		a.log.ErrorContext(ctx, fmt.Sprintf("%s: wrong type of usecase error", component), slog.Any("error", err))
		return http.StatusInternalServerError, "internal error"
	}

	if code >= 500 {
		a.log.ErrorContext(ctx, msg, slog.Any("error", err))
		return code, "internal error"
	}

	a.log.WarnContext(ctx, msg)
	return code, strings.Split(msg, ": ")[1]
}

func (a *Adapter) parseRecieveFileError(ctx context.Context, component string, err error) (int, string) {
	var recieveError *infra.RecieveFileError
	if errors.As(err, &recieveError) {
		var msg string

		if recieveError.Code == http.StatusInternalServerError {
			a.log.ErrorContext(ctx, fmt.Sprintf("%s: multipart error", component), slog.Any("error", err))
			msg = "internal error"
		} else {
			msg = recieveError.Error()
		}

		return recieveError.Code, msg
	}

	a.log.ErrorContext(ctx, fmt.Sprintf("%s: wrong type of multipart error", component), slog.String("error", "error expected to be recieveFileError type"))
	return http.StatusInternalServerError, "internal error"
}
