package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const serviceName = "fiber-v3-template"

// Setup configures OpenTelemetry providers and returns a shutdown function.
func Setup(ctx context.Context, endpoint string) (func(context.Context) error, error) {
	cleanEndpoint := strings.TrimSpace(endpoint)
	if cleanEndpoint == "" {
		tracerProvider := trace.NewTracerProvider()
		meterProvider := metric.NewMeterProvider()
		otel.SetTracerProvider(tracerProvider)
		otel.SetMeterProvider(meterProvider)

		return func(context.Context) error {
			return nil
		}, nil
	}

	// Single resource with one semconv schema URL — do not merge with resource.Default()
	// (Default uses a newer schema and triggers "conflicting Schema URL" with semconv v1.26).
	telemetryResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("environment", "runtime"),
	)

	traceExporter, traceExporterError := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(cleanEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if traceExporterError != nil {
		return nil, fmt.Errorf("create trace exporter: %w", traceExporterError)
	}

	metricExporter, metricExporterError := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(cleanEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if metricExporterError != nil {
		return nil, fmt.Errorf("create metric exporter: %w", metricExporterError)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(telemetryResource),
		trace.WithBatcher(traceExporter),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(telemetryResource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)

	return func(shutdownContext context.Context) error {
		tracerShutdownError := tracerProvider.Shutdown(shutdownContext)
		meterShutdownError := meterProvider.Shutdown(shutdownContext)
		if tracerShutdownError != nil && meterShutdownError != nil {
			return fmt.Errorf("shutdown tracer provider: %w; shutdown meter provider: %w", tracerShutdownError, meterShutdownError)
		}
		if tracerShutdownError != nil {
			return fmt.Errorf("shutdown tracer provider: %w", tracerShutdownError)
		}
		if meterShutdownError != nil {
			return fmt.Errorf("shutdown meter provider: %w", meterShutdownError)
		}

		return nil
	}, nil
}
