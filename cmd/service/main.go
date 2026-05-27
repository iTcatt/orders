package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"iTcatt/orders/internal/api"
	apiCategory "iTcatt/orders/internal/api/category"
	apiImage "iTcatt/orders/internal/api/image"
	apiProduct "iTcatt/orders/internal/api/product"
	minioInfra "iTcatt/orders/internal/infra/minio"
	"iTcatt/orders/internal/infra/postgres"
	storageCategories "iTcatt/orders/internal/storage/categories"
	storageImages "iTcatt/orders/internal/storage/images"
	storageObjects "iTcatt/orders/internal/storage/objects"
	"iTcatt/orders/internal/storage/products"
	categoryUsecase "iTcatt/orders/internal/usecase/category"
	imageUsecase "iTcatt/orders/internal/usecase/image"
	productUsecase "iTcatt/orders/internal/usecase/product"
	"iTcatt/orders/pkg/sqlp"
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

	godotenv.Load() //nolint:errcheck,gosec // .env is optional, for local dev only

	db, err := postgres.New(os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			slog.Error("close database", slog.Any("error", err))
		}
	}()

	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	minioClient, err := minioInfra.New(minioInfra.Config{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		UseSSL:    useSSL,
	})
	if err != nil {
		return fmt.Errorf("create minio client: %w", err)
	}

	objectStore := storageObjects.New(minioClient, os.Getenv("MINIO_BUCKET"), os.Getenv("MINIO_PUBLIC_URL"))
	if err := objectStore.EnsureBucket(ctx); err != nil {
		return fmt.Errorf("ensure minio bucket: %w", err)
	}

	idGen := func() string { return uuid.Must(uuid.NewV7()).String() }

	productDB := products.New(db)
	imageDB := storageImages.New(db)
	categoryDB := storageCategories.New(db)

	productUC := productUsecase.New(productDB, imageDB, time.Now, idGen)
	txManager := sqlp.NewTxManager(db)
	imageUC := imageUsecase.New(imageDB, objectStore, productDB, txManager, time.Now, idGen)
	categoryUC := categoryUsecase.New(categoryDB)

	productHandler := apiProduct.New(productUC)
	imageHandler := apiImage.New(imageUC)
	categoryHandler := apiCategory.New(categoryUC)

	router := api.NewRouter(productHandler, imageHandler, categoryHandler)

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
