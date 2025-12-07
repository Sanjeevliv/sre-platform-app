package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sanjeevsethi/sre-platform-app/internal/config"
	"github.com/sanjeevsethi/sre-platform-app/internal/worker"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 7. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	// Check Redis connection
	// Create a context that we can cancel to signal shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis at %s: %v", cfg.RedisAddr, err)
	}
	log.Printf("Connected to Redis at %s", cfg.RedisAddr)

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
		log.Printf("Starting worker-service metrics server on :%s...", cfg.WorkerPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Listen for interrupts
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown
	log.Println("Shutdown signal received...")

	// 1. Signal worker to stop
	cancel()

	// 2. Wait for worker to finish
	log.Println("Waiting for worker to exit...")
	wg.Wait()
	log.Println("Worker exited.")

	// 3. Shutdown metrics server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server forced to shutdown: %v", err)
	}
	log.Println("Metrics server stopped.")
}
