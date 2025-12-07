package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
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

// Job represents the work unit with metadata
type Job struct {
	ID        string `json:"id"`
	Payload   string `json:"payload"`
	RequestID string `json:"request_id"`
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

		// 4. Simulate job processing
		start := time.Now()
		time.Sleep(100 * time.Millisecond)
		duration := time.Since(start).Seconds()

		// 5. Instrument the outcome
		status := "success"
		if job.Payload == "fail_me" {
			status = "error"
			jobsFailedTotal.Inc()
			l.Warn().Str("payload", job.Payload).Msg("Job failed")
		} else {
			jobsProcessedTotal.Inc()
		}
		jobDuration.WithLabelValues(status).Observe(duration)
	}
}
