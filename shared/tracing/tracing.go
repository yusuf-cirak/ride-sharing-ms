package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
}

func InitTracer(cfg Config) (func(context.Context) error, error) {
	// Exporter
	traceExporter, err := newExporter(cfg.JaegerEndpoint)
	if err != nil {
		return nil, err
	}

	// Trace Provider
	traceProvider, err := newTraceProvider(cfg, traceExporter)

	if traceProvider == nil {
		return nil, err
	}

	otel.SetTracerProvider(traceProvider)

	// Propagator
	otel.SetTextMapPropagator(newPropagator())

	return traceProvider.Shutdown, nil
}

func newExporter(endpoint string) (sdktrace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(cfg Config, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String((cfg.ServiceName)),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment)))

	if err != nil {
		return nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))

	if err != nil {
		return nil, err
	}

	return traceProvider, nil
}
