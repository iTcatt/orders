package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/joho/godotenv"

	"iTcatt/orders/internal/api"
	apiProduct "iTcatt/orders/internal/api/product"
	"iTcatt/orders/internal/infra/postgres"
	"iTcatt/orders/internal/storage/products"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	setupLogger()

	db, err := postgres.New(os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			slog.Error("close error", slog.String("err", err.Error()))
		}
	}()

	productDB := products.New(db)

	productHandler := apiProduct.New(productDB)

	router := api.NewRouter(productHandler)

	server := &http.Server{
		Addr:         ":8081",
		IdleTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      router,
	}

	go func() {
		slog.Info("Start server")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	slog.Info("Shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
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
