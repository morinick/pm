package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"passman/cmd/internal/infra"
	"passman/cmd/internal/users"
	"strings"

	"github.com/go-chi/chi/v5"
	vldtr "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Adapter struct {
	log     *slog.Logger
	uu      userUsecase
	session sessionManager
	v       *validator
}

func NewRouter(ua userUsecase, sm sessionManager, v *vldtr.Validate) chi.Router {
	a := &Adapter{
		log:     slog.Default(),
		uu:      ua,
		session: sm,
		v:       newValidator(v),
	}

	router := chi.NewRouter()
	router.Post("/registration", a.Registration)
	router.Post("/login", a.Login)
	router.Delete("/logout", a.Logout)

	routerAuth := chi.NewRouter()
	routerAuth.Use(infra.AuthMiddleware(sm))
	routerAuth.Put("/update", a.UpdateUser)
	routerAuth.Delete("/delete", a.DeleteUser)

	router.Mount("/", routerAuth)

	return router
}

func (a *Adapter) Registration(w http.ResponseWriter, r *http.Request) {
	candidate := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
		a.log.ErrorContext(r.Context(), "Registration: can't parse body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateUserCreds(candidate.Username, candidate.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := a.uu.Registration(
		r.Context(),
		users.UserDTO{
			Username: candidate.Username,
			Password: candidate.Password,
		},
	)
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "Registration", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	a.session.Put(r.Context(), "user_id", userID.String())

	infra.ResponseJSON(w, struct {
		UserID string `json:"user_id"`
	}{UserID: userID.String()}, http.StatusOK)
}

func (a *Adapter) Login(w http.ResponseWriter, r *http.Request) {
	candidate := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
		a.log.ErrorContext(r.Context(), "Login: can't parse body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateUserCreds(candidate.Username, candidate.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := a.uu.Login(
		r.Context(),
		users.UserDTO{
			Username: candidate.Username,
			Password: candidate.Password,
		},
	)
	if err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "Login", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	a.session.Put(r.Context(), "user_id", userID.String())

	infra.ResponseJSON(w, struct {
		UserID string `json:"user_id"`
	}{UserID: userID.String()}, http.StatusOK)
}

func (a *Adapter) Logout(w http.ResponseWriter, r *http.Request) {
	if err := a.session.Destroy(r.Context()); err != nil {
		a.log.ErrorContext(r.Context(), "Logout: can't destroy session", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.session.GetString(r.Context(), "user_id"))

	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "UpdateUser: can't parse body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateUserCreds(body.Username, body.Password); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.uu.UpdateUser(r.Context(), userID, users.UserDTO{
		Username: body.Username,
		Password: body.Password,
	}); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdateUser", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.session.GetString(r.Context(), "user_id"))

	if err := a.uu.DeleteUser(r.Context(), userID); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "DeleteUser", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	if err := a.session.Destroy(r.Context()); err != nil {
		a.log.ErrorContext(r.Context(), "DeleteUser: can't destroy session", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) ParseUsecaseError(ctx context.Context, component string, usecaseError error) (int, string) {
	code, msg, err := a.uu.ParseUserError(usecaseError)
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
