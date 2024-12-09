package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func SetUpMetrics(ctx context.Context, serviceName string) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, err
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resources),
		sdkmetric.WithReader(
			// collects and exports metric data every 30 seconds.
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)

	otel.SetMeterProvider(mp)

	return mp, nil

}
