package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
)

type Job struct {
	ID        string `json:"id"`
	Payload   string `json:"payload"`
	RequestID string `json:"request_id"`
}

type Producer struct {
	client *redis.Client
	cb     *gobreaker.CircuitBreaker
}

func NewProducer(addr string) *Producer {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	// Enable tracing
	rdb.AddHook(redisotel.NewTracingHook())

	st := gobreaker.Settings{
		Name:        "Redis",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)

	return &Producer{
		client: rdb,
		cb:     cb,
	}
}

func (p *Producer) Enqueue(ctx context.Context, job Job) error {
	_, err := p.cb.Execute(func() (interface{}, error) {
		data, err := json.Marshal(job)
		if err != nil {
			return nil, err
		}
		return p.client.LPush(ctx, "jobs", data).Result()
	})
	if err != nil {
		return fmt.Errorf("enqueue failed: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.client.Close()
}
