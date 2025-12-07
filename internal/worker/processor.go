package worker

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

// Start begins the worker processing loop. It blocks until the context is done.
func Start(ctx context.Context, rdb *redis.Client) {
	log.Println("Starting worker process loop...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done, stopping worker loop.")
			return
		default:
			// Continue
		}

		// 3. Block and wait for a new job on the 'jobs' queue
		// BRPop blocks until a job is available or a timeout occurs
		result, err := rdb.BRPop(ctx, 1*time.Second, "jobs").Result()
		if err != nil {
			if err == redis.Nil {
				// Timeout, just continue
				continue
			}
			// Don't count Redis timeouts/nil results as a "failed" job
			// log.Printf("Error pulling from Redis queue: %v", err) // Optional: reduce log spam
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
