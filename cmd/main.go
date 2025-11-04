package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	credsDB "passman/cmd/internal/creds/adapters/db"
	credsHTTP "passman/cmd/internal/creds/adapters/http"
	credsUsecases "passman/cmd/internal/creds/usecases"
	usersDB "passman/cmd/internal/users/adapters/db"
	usersHTTP "passman/cmd/internal/users/adapters/http"
	usersUsecases "passman/cmd/internal/users/usecases"
	"passman/internal/backups"
	database "passman/pkg/database/sqlite"
	"passman/pkg/logger"
	"passman/pkg/session"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/traceid"
	"github.com/go-playground/validator/v10"
	"golang.org/x/sync/errgroup"
)

func main() {
	mainCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flags := ParseFlags()

	cfg, err := newConfig()
	if err != nil {
		slog.Error("Configuration error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	sm, err := session.NewSessionManager(mainCtx, session.SessionManagerOptions{CookieHTTPOnly: true})
	if err != nil {
		slog.Error("Session error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	dbStorage, err := database.NewDB(mainCtx, "data.db")
	if err != nil {
		slog.Error("Database error", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbStorage.Close()

	backupController, err := backups.New("data.db", "/backup", flags["key"])
	if err != nil {
		slog.Error("Backup controller error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// If saved_database.bak exist, then ciphers have already been read from it
	if backupController.Ciphers == nil {
		ciphers, err := backups.GenerateCiphers(10, 32)
		if err != nil {
			slog.Error("Secrets generation error", slog.String("error", err.Error()))
			os.Exit(1)
		}
		backupController.Ciphers = ciphers
	}

	logger.SetNewLoggerByDefault(cfg.LogLevel)

	globalValidator := validator.New(validator.WithRequiredStructEnabled())

	appRouter := chi.NewRouter()
	appRouter.Use(
		middleware.Heartbeat("/healthcheck"),
		middleware.Timeout(30*time.Second),
		traceid.Middleware,
		logger.HTTPLogger(cfg.LogLevel),
		sm.LoadAndSave,
	)

	// User domain
	userDBRepository := usersDB.New(dbStorage)
	userUsecase := usersUsecases.New(userDBRepository)
	userRouter := usersHTTP.NewRouter(userUsecase, sm, globalValidator)
	appRouter.Mount("/user", userRouter)

	// Creds domain
	credsRepository := credsDB.New(dbStorage)
	credsUsecase := credsUsecases.New(credsRepository, backupController.Ciphers)
	credsRouter := credsHTTP.NewRouter(credsUsecase, sm, globalValidator)
	appRouter.Mount("/creds", credsRouter)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: appRouter,
	}

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		slog.Default().Info("Server started", slog.String("address", srv.Addr))
		return srv.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()

		sCtx, sCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer sCancel()

		if err := srv.Shutdown(sCtx); err != nil {
			return err
		}

		if err := backupController.SaveDatabase("saved_database.bak"); err != nil {
			return err
		}
		slog.Default().Info("Database encrypted", slog.String("key", backupController.DecryptKey))

		slog.Default().Info("Successful shutdown")
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
