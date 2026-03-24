package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/joho/godotenv"

	"iTcatt/orders/internal/api"
	apiProduct "iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/infra/postgres"
	"iTcatt/orders/internal/storage/products"
	productUsecase "iTcatt/orders/internal/usecase/product"
)

func main() {
	setupLogger()
	if err := run(); err != nil {
		slog.Error("service stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf(".env file not loaded: %w", err)
	}

	db, err := postgres.New(os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			slog.Error("close error", slog.String("err", err.Error()))
		}
	}()

	productDB := products.New(db)
	productUC := productUsecase.New(productDB, time.Now, rand.Uint32)
	productHandler := apiProduct.New(productUC)

	router := api.NewRouter(productHandler)

	server := &http.Server{
		Addr:         ":8081",
		IdleTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      router,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("Start server")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case <-ch:
	}

	slog.Info("Shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

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
