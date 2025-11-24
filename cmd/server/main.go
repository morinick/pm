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

	accountsDB "passman/internal/server/accounts/adapters/db"
	accountsHTTP "passman/internal/server/accounts/adapters/http"
	accountsUsecases "passman/internal/server/accounts/usecases"
	"passman/internal/server/backups"
	servicesDB "passman/internal/server/services/adapters/db"
	servicesHTTP "passman/internal/server/services/adapters/http"
	servicesUsecases "passman/internal/server/services/usecases"
	"passman/internal/server/starter"
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

	dbStorage, err := database.NewDB(mainCtx, cfg.DBURL)
	if err != nil {
		slog.Error("Database error", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbStorage.Close()

	m, err := migrator.NewMigrator("file:///migrations", cfg.DBURL)
	if err != nil {
		slog.Error("Migrator error", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := m.Up(); err != nil {
		slog.Error("Up migrations error", slog.String("error", err.Error()))
	}
	m.Close()

	backupController := backups.New(
		backups.ControllerOptions{
			DBURL:     cfg.DBURL,
			BackupDir: cfg.BackupDir,
			AssetsDir: cfg.AssetsDir,
			MasterKey: cfg.MasterKey,
		},
	)

	ciphers, err := starter.Start(
		mainCtx,
		starter.StartOptions{
			DB:               dbStorage,
			BackupController: backupController,
			AssetsDir:        cfg.AssetsDir,
			MasterKey:        cfg.MasterKey,
		},
	)
	if err != nil {
		slog.Error("Starter error", slog.String("error", err.Error()))
		os.Exit(1)
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
	userRepository := usersDB.New(dbStorage)
	userUsecase := usersUsecases.New(userRepository)
	userRouter := usersHTTP.NewRouter(userUsecase, sm, globalValidator)
	appRouter.Mount("/users", userRouter)

	// Accounts domain
	accountsRepository := accountsDB.New(dbStorage)
	accountsUsecase := accountsUsecases.New(accountsRepository, ciphers)
	accountsRouter := accountsHTTP.NewRouter(accountsUsecase, sm, globalValidator)
	appRouter.Mount("/accounts", accountsRouter)

	// Services domain
	servicesRepository := servicesDB.New(dbStorage)
	servicesUsecase := servicesUsecases.New(servicesRepository, cfg.AssetsDir)
	servicesRouter := servicesHTTP.NewRouter(servicesUsecase, sm, globalValidator)
	appRouter.Mount("/services", servicesRouter)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: appRouter,
	}

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		if len(cfg.MasterKey) == 0 {
			slog.Default().Info("It's first starting", slog.String("MasterKey", backupController.Key))
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

		slog.Default().Info("Backup created")

		if err := saveMasterKey(backupController.Key); err != nil {
			slog.Default().Warn("Failed saving master key", slog.String("error", err.Error()))
		}

		slog.Default().Info("Successful shutdown")
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
