package tracing

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// // NewJaegerProvider returns a new jaeger-based tracing provider.
// func NewJaegerProvider(url string, serviceName string) (*tracesdk.TracerProvider, error) {
// 	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	tp := tracesdk.NewTracerProvider(
// 		tracesdk.WithBatcher(exp),
// 		tracesdk.WithResource(resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceNameKey.String(serviceName),
// 		)),
// 	)
// 	return tp, nil
// }

// http tracing
func NewHTTPTracerProvider(url, serviceName string) (*tracesdk.TracerProvider, error) {
	exp, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(url))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return tp, nil
}

// grpc tracing
func NewGrpcTracerProvider(url, serviceName string) (*tracesdk.TracerProvider, error) {
	if url == "" {
		return nil, errors.New("jaegarURL is empty")
	}
	exp, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(url))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return tp, nil
}
