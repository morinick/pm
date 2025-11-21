package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"passman/internal/server/infra"
	"passman/internal/server/users"

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

	routerAuth.Put("/update/username", a.UpdateUsername)
	routerAuth.Put("/update/password", a.UpdatePassword)
	routerAuth.Delete("/delete", a.RemoveUser)

	router.Mount("/", routerAuth)

	return router
}

func (a *Adapter) Registration(w http.ResponseWriter, r *http.Request) {
	candidate := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
		a.log.ErrorContext(r.Context(), "Registration: failed parsing body", slog.Any("error", err))
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
		a.log.ErrorContext(r.Context(), "Login: failed parsing body", slog.Any("error", err))
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
		a.log.ErrorContext(r.Context(), "Logout: failed destroying session", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.session.GetString(r.Context(), "user_id"))

	body := struct {
		OldPassword string `json:"old_password"`
		Username    string `json:"username"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "UpdateUsername: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidateUserCreds(body.Username, body.OldPassword); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser := users.UpdatedUserParams{
		UserID:      userID,
		OldPassword: body.OldPassword,
		UserDTO: users.UserDTO{
			Username: body.Username,
		},
	}

	if err := a.uu.UpdateUser(r.Context(), updatedUser); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdateUsername", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.session.GetString(r.Context(), "user_id"))

	body := struct {
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.log.ErrorContext(r.Context(), "UpdatePassword: failed parsing body", slog.Any("error", err))
		infra.ErrorHandler(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := a.v.ValidatePasswords(body.Password, body.OldPassword); err != nil {
		infra.ErrorHandler(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser := users.UpdatedUserParams{
		UserID:      userID,
		OldPassword: body.OldPassword,
		UserDTO: users.UserDTO{
			Password: body.Password,
		},
	}

	if err := a.uu.UpdateUser(r.Context(), updatedUser); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "UpdatePassword", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Adapter) RemoveUser(w http.ResponseWriter, r *http.Request) {
	userID := uuid.MustParse(a.session.GetString(r.Context(), "user_id"))

	if err := a.uu.RemoveUser(r.Context(), userID); err != nil {
		code, msg := a.ParseUsecaseError(r.Context(), "DeleteUser", err)
		infra.ErrorHandler(w, code, msg)
		return
	}

	if err := a.session.Destroy(r.Context()); err != nil {
		a.log.ErrorContext(r.Context(), "RemoveUser: failed destroying body", slog.Any("error", err))
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
