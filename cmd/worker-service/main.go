package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	stdlog "log"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/sanjeevsethi/sre-platform-app/internal/config"
	"github.com/sanjeevsethi/sre-platform-app/internal/logger"
	"github.com/sanjeevsethi/sre-platform-app/internal/telemetry"
	"github.com/sanjeevsethi/sre-platform-app/internal/worker"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	logger.Init("info", os.Getenv("GIN_MODE") != "release")

	// 3. Initialize Tracing
	shutdownTracer, err := telemetry.InitTracer("worker-service")
	if err != nil {
		stdlog.Printf("Failed to init tracer: %v", err)
	} else {
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				stdlog.Printf("Error shutting down tracer: %v", err)
			}
		}()
	}

	// 7. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	// Add Redis instrumentation hook
	rdb.AddHook(redisotel.NewTracingHook())

	// Check Redis connection
	// Create a context that we can cancel to signal shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal().Err(err).Str("addr", cfg.RedisAddr).Msg("Unable to connect to Redis")
	}
	log.Info().Str("addr", cfg.RedisAddr).Msg("Connected to Redis")

	// 8. Launch the worker loop in a background goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx, rdb)
	}()

	// 9. Expose /metrics for Prometheus
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	metricsSrv := &http.Server{
		Addr:    ":" + cfg.WorkerPort,
		Handler: mux,
	}

	go func() {
		log.Info().Str("port", cfg.WorkerPort).Msg("Starting worker-service metrics server")
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Metrics server error")
		}
	}()

	// Listen for interrupts
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown
	log.Info().Msg("Shutdown signal received...")

	// 1. Signal worker to stop
	cancel()

	// 2. Wait for worker to finish
	log.Info().Msg("Waiting for worker to exit...")
	wg.Wait()
	log.Info().Msg("Worker exited.")

	// 3. Shutdown metrics server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Metrics server forced to shutdown")
	}
	log.Info().Msg("Metrics server stopped.")
}
