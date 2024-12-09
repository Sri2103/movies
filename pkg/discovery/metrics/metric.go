package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
)

func InitMetrics(ctx context.Context, service string) (*metricSdk.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	provider := metricSdk.NewMeterProvider(metricSdk.WithReader(exporter))

	otel.SetMeterProvider(provider)

	return provider, nil
}

func StartMetricsEndPoint(port int) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting metrics endpoint on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
