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

	"passman/internal/server/backups"
	credsDB "passman/internal/server/creds/adapters/db"
	credsHTTP "passman/internal/server/creds/adapters/http"
	credsUsecases "passman/internal/server/creds/usecases"
	servicesDB "passman/internal/server/services/adapters/db"
	servicesHTTP "passman/internal/server/services/adapters/http"
	servicesUsecases "passman/internal/server/services/usecases"
	usersDB "passman/internal/server/users/adapters/db"
	usersHTTP "passman/internal/server/users/adapters/http"
	usersUsecases "passman/internal/server/users/usecases"
	"passman/pkg/database/migrator"
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

	backupController := backups.New("data.db", "/backup", flags["key"])

	if err := backupController.LoadBackup(); err != nil && !errors.Is(err, backups.ErrEmptyKey) {
		slog.Error("Backup controller error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	m, err := migrator.NewMigrator("file:///migrations", "data.db")
	if err != nil {
		slog.Error("Migrator error", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := m.Up(); err != nil {
		slog.Error("Up migrations error", slog.String("error", err.Error()))
	}
	m.Close()

	// Means first start
	firstStartFlag := false // to output a message to the logger
	if len(backupController.Key) == 0 {
		firstStartFlag = true
		ciphers, err := backups.GenerateCiphers(10)
		if err != nil {
			slog.Error("Secrets generation error", slog.String("error", err.Error()))
			os.Exit(1)
		}
		backupController.Ciphers = ciphers

		if err := backups.InsertAssetsToDB(dbStorage); err != nil {
			slog.Error("Inserting assets error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	logger.SetNewLoggerByDefault(cfg.LogLevel)

	globalValidator := validator.New(validator.WithRequiredStructEnabled())

	appRouter := chi.NewRouter()
	appRouter.Use(
		middleware.Timeout(30*time.Second),
		traceid.Middleware,
		logger.HTTPLogger(cfg.LogLevel),
		sm.LoadAndSave,
	)

	// Users domain
	userDBRepository := usersDB.New(dbStorage)
	userUsecase := usersUsecases.New(userDBRepository)
	userRouter := usersHTTP.NewRouter(userUsecase, sm, globalValidator)
	appRouter.Mount("/users", userRouter)

	// Creds domain
	credsRepository := credsDB.New(dbStorage)
	credsUsecase := credsUsecases.New(credsRepository, backupController.Ciphers)
	credsRouter := credsHTTP.NewRouter(credsUsecase, sm, globalValidator)
	appRouter.Mount("/creds", credsRouter)

	// Services domain
	servicesRepository := servicesDB.New(dbStorage)
	servicesUsecase := servicesUsecases.New(servicesRepository, "/assets")
	servicesRouter := servicesHTTP.NewRouter(servicesUsecase, sm, globalValidator)
	appRouter.Mount("/services", servicesRouter)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: appRouter,
	}

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		if firstStartFlag {
			slog.Default().Info("First starting...")
		}
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

		if err := backupController.SaveBackup(); err != nil {
			return err
		}
		slog.Default().Info("Backup created", slog.String("key", backupController.Key))

		slog.Default().Info("Successful shutdown")
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
