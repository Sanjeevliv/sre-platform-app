package main

import (
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sanjeevsethi/sre-platform-app/internal/api"
	"github.com/sanjeevsethi/sre-platform-app/internal/config"
	"github.com/sanjeevsethi/sre-platform-app/internal/logger"
	"github.com/sanjeevsethi/sre-platform-app/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatalf("Failed to load config: %v", err)
	}

	// 3. Initialize Logger
	// In production, we'd probably want this to be false (JSON logs)
	// For dev, reading console logs is nicer.
	// We could put this in config too: cfg.LogPretty
	logger.Init("info", os.Getenv("GIN_MODE") != "release")

	// 4. Initialize Tracing
	shutdownTracer, err := telemetry.InitTracer("api-service")
	if err != nil {
		stdlog.Printf("Failed to init tracer: %v", err)
		// We don't fatal here to allow running without collector in dev if needed,
		// though strictly SRE practice says observability is critical.
	} else {
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				stdlog.Printf("Error shutting down tracer: %v", err)
			}
		}()
	}

	// Get the configured mux from the internal package
	r := api.NewServer()

	// Register Middleware
	r.Use(otelgin.Middleware("api-service")) // OpenTelemetry
	r.Use(api.RequestIDMiddleware())
	r.Use(api.LoggerMiddleware())

	srv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: r,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	go func() {
		log.Info().Str("port", cfg.APIPort).Msg("Starting api_service")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	select {
	case err := <-serverErrors:
		log.Fatal().Err(err).Msg("Error starting server")

	case <-shutdown:
		log.Info().Msg("Start shutdown...")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Dur("timeout", 5*time.Second).Msg("Graceful shutdown did not complete")
			if err := srv.Close(); err != nil {
				log.Fatal().Err(err).Msg("Could not stop http server")
			}
		}
	}
	log.Info().Msg("Server stopped")
}
