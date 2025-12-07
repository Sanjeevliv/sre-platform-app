package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer configures an OpenTelemetry exporter and tracer provider.
// It returns a shutdown function that should be called on service exit.
func InitTracer(serviceName string) (func(context.Context) error, error) {
	// 1. Create OTLP Exporter (HTTP)
	// Uses OTEL_EXPORTER_OTLP_ENDPOINT env var if set (default localhost:4318)
	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		return nil, err
	}

	// 2. Identify the resource (service name, etc)
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			// semconv.ServiceVersionKey.String(metadata.Version), // Avoid circular dep for now
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. Create Tracer Provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// 4. Set Global Provider and Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
