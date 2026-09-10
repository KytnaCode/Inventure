// Package main contains application's entry point.
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kytnacode/inventure/event"
	"github.com/kytnacode/inventure/internal/auth"
	"github.com/kytnacode/inventure/internal/config"
	"github.com/kytnacode/inventure/internal/database"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/routes"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/logging"
	"github.com/kytnacode/inventure/server"
	"github.com/kytnacode/inventure/validation"
)

const (
	shutdownTimeout    = time.Minute
	readHeaderTimeout  = time.Second * 5
	sessionIdleTimeout = time.Hour
)

func main() {
	if ok := run(); !ok {
		os.Exit(1)
	}
}

func run() bool {
	conf, err := config.Read()
	if err != nil {
		logging.New(false).Error("could not read config", logging.Error(err))

		return false
	}

	logger := logging.New(conf.Debug)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx = logging.WithLogger(ctx, logger)

	sessionManager := auth.NewManager(&auth.SessionStoreConfig{
		IdleTimeout: sessionIdleTimeout,
	})

	v := validation.New()

	logger.Info("trying to connect to database")

	db, err := database.Connect(ctx, &conf.Database)
	if err != nil {
		logger.Error("could not connect to database", logging.Error(err))

		return false
	}

	logger.Info("connected to database")

	logger.Info("running database migrations")

	err = database.RunMigrations(db)
	if err != nil {
		logger.Error("could not run database migrations", logging.Error(err))

		return false
	}

	logger.Info("databse migrations ended succefully")

	br := event.NewBroker()

	logger.Info("starting event broker")

	go br.Run(ctx)

	retailRepository := retail.NewRepository(db)

	retailService := retail.NewService(retail.NewServiceProvider(db), retailRepository)

	retailEvents := retail.NewEventHandler(retailService, br)

	logger.Info("registering app events")

	go retailEvents.Handle(ctx)

	userRepo := user.NewRepository(db)

	userService := user.NewService(userRepo)

	h := routes.SetupRouter(&routes.Config{
		LoggerMiddleware:         web.NewLoggerMiddleware(logger, conf.Debug),
		IPMiddleware:             web.IPMiddleware(conf.API.TrustedProxies),
		EmbeddedLoggerMiddelware: web.WithEmbeddedLogger(logger),
		SessionManager:           sessionManager,
		AuthRoutes: auth.NewRoutes(&auth.RoutesConfig{
			SessionManager:        sessionManager,
			LoginAttemptLimit:     conf.API.LoginAttempt.RequestLimit,
			LoginAttempTimeWindow: time.Duration(conf.API.LoginAttempt.TimeWindowSeconds) * time.Second,
			RequestLimit:          conf.API.PasswordAuth.RequestLimit,
			TimeWindow:            time.Duration(conf.API.LoginAttempt.TimeWindowSeconds) * time.Second,
			Validator:             v,
			UserService:           userService,
			Broker:                br,
		}),
	})

	serv := http.Server{
		Addr:              conf.Server.Addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}

	logger.Info("starting server")

	err = server.ListenAndServe(ctx, &server.Config{
		ShutdownTimeout: shutdownTimeout,
		Server:          &serv,
		CertFile:        conf.Server.CertFile,
		KeyFile:         conf.Server.KeyFile,
	})
	if err != nil {
		logger.Error("server error", logging.Error(err))

		return false
	}

	logger.Info("stopping server")

	return true
}
