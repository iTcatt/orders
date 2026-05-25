package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"iTcatt/orders/internal/api"
	apiProduct "iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/infra/postgres"
	"iTcatt/orders/internal/storage/products"
	productUsecase "iTcatt/orders/internal/usecase/product"
)

const shutdownTimeout = 5 * time.Second

func main() {
	setupLogger()
	if err := run(); err != nil {
		slog.Error("service stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	godotenv.Load() //nolint:errcheck // .env is optional, for local dev only

	db, err := postgres.New(os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			slog.Error("close database", slog.Any("error", err))
		}
	}()

	productDB := products.New(db)

	productUC := productUsecase.New(productDB, time.Now, func() string {
		return uuid.Must(uuid.NewV7()).String()
	})
	productHandler := apiProduct.New(productUC)

	router := api.NewRouter(productHandler)

	server := &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler:           router,
	}

	go func() {
		slog.Info("Start server")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", slog.Any("error", err))
		}
	}()
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	slog.Info("Shutdown server")
	return nil
}

func setupLogger() {
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	var logger *slog.Logger
	if os.Getenv("env") == "prod" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, slogOpts))
	} else {
		opts := &devslog.Options{
			HandlerOptions:    slogOpts,
			MaxSlicePrintSize: 10,
			SortKeys:          true,
			NewLineAfterLog:   true,
			StringerFormatter: true,
		}

		logger = slog.New(devslog.NewHandler(os.Stdout, opts))
	}

	slog.SetDefault(logger)
}
