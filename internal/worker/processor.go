package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sanjeevsethi/sre-platform-app/internal/queue"
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
	queueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_queue_depth",
			Help: "Current depth of the jobs queue in Redis.",
		},
	)
	jobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_job_duration_seconds",
			Help:    "Duration of job processing.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

// Use shared Job struct from queue package
type Job = queue.Job

const MaxRetries = 3

func processJobWithRetry(l zerolog.Logger, job Job) (string, error) {
	var err error
	for i := 0; i <= MaxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<i) * 100 * time.Millisecond // 200ms, 400ms, 800ms
			l.Warn().Int("attempt", i+1).Dur("backoff", backoff).Msg("Retrying job...")
			time.Sleep(backoff)
		}

		// 4. Simulate job processing
		start := time.Now()
		// Logic:
		// Check payload. If "fail_me", simulate error.
		// If "fail_once", simulate error only on first attempt.
		status := "success"
		err = nil

		if job.Payload == "fail_me" {
			err = fmt.Errorf("simulated permanent failure")
		} else if job.Payload == "fail_once" && i == 0 {
			err = fmt.Errorf("simulated transient failure")
		} else {
			time.Sleep(100 * time.Millisecond)
		}

		duration := time.Since(start).Seconds()

		if err != nil {
			status = "error"
		} else {
			jobDuration.WithLabelValues(status).Observe(duration)
			return status, nil // Success
		}
	}
	return "error", err // Retries exhausted
}

// Start begins the worker processing loop. It blocks until the context is done.
func Start(ctx context.Context, rdb *redis.Client) {
	log.Info().Msg("Starting worker process loop...")

	// Launch background monitor for queue depth
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				val, err := rdb.LLen(ctx, "jobs").Result()
				if err == nil {
					queueDepth.Set(float64(val))
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context done, stopping worker loop.")
			return
		default:
			// Continue
		}

		// 3. Block and wait for a new job on the 'jobs' queue
		// BRPop blocks until a job is available or a timeout occurs
		result, err := rdb.BRPop(ctx, 1*time.Second, "jobs").Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			// log.Printf("Error pulling from Redis queue: %v", err)
			continue
		}

		// result[1] is the job data (JSON string)
		rawJob := result[1]

		var job Job
		// Try parsing as JSON. If fails, assume legacy string format.
		if err := json.Unmarshal([]byte(rawJob), &job); err != nil {
			// Handle legacy string jobs or malformed JSON
			job = Job{
				ID:        "legacy",
				Payload:   rawJob,
				RequestID: "unknown",
			}
		}

		// Create a logger with context for this job
		l := log.With().Str("job_id", job.ID).Str("request_id", job.RequestID).Logger()
		l.Info().Str("payload", job.Payload).Msg("Processing job")

		status, err := processJobWithRetry(l, job)

		if err != nil {
			jobsFailedTotal.Inc()
			l.Error().Err(err).Msg("Job failed after retries")
			// Future: Push to Dead Letter Queue (DLQ)
		} else {
			jobsProcessedTotal.Inc()
			l.Info().Str("status", status).Msg("Job processed successfully")
		}
	}
}
