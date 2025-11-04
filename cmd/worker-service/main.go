package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 1. Define Prometheus metrics for the worker
var (
	jobsProcessedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_service_jobs_processed_total",
			Help: "Total number of jobs processed by the worker.",
		},
	)
	jobsFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_service_jobs_failed_total",
			Help: "Total number of jobs that failed processing.",
		},
	)
)

// 2. processJobs is our main worker loop
func processJobs(ctx context.Context, rdb *redis.Client) {
	log.Println("Starting worker process loop...")

	for {
		// 3. Block and wait for a new job on the 'jobs' queue
		// BRPop blocks until a job is available or a timeout occurs
		result, err := rdb.BRPop(ctx, 0, "jobs").Result()
		if err != nil {
			// Don't count Redis timeouts/nil results as a "failed" job
			log.Printf("Error pulling from Redis queue: %v", err)
			continue
		}

		// result is a []string, where result[0] is the queue name and result[1] is the job data
		jobData := result[1]
		log.Printf("Processing job: %s", jobData)

		// 4. Simulate job processing
		time.Sleep(100 * time.Millisecond) // Simulate work

		// 5. Instrument the outcome
		// In a real app, you'd check for processing errors here.
		// For now, we'll just assume success.
		if jobData == "fail_me" {
			jobsFailedTotal.Inc()
		} else {
			jobsProcessedTotal.Inc()
		}
	}
}

func main() {
	// 6. Get Redis address from Environment Variable, or use default
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local dev
	}

	// 7. Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Check Redis connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis at %s: %v", redisAddr, err)
	}
	log.Printf("Connected to Redis at %s", redisAddr)

	// 8. Launch the worker loop in a background goroutine
	go processJobs(ctx, rdb)

	// 9. Expose /metrics for Prometheus
	http.Handle("/metrics", promhttp.Handler())

	// We also expose /healthz for Kubernetes liveness probes
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("Starting worker-service metrics server on :8081...")
	// We run this on a different port (:8081) so it can run
	// locally alongside the api-service (:8080)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
}
