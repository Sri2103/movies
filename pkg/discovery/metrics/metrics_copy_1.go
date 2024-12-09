package metrics

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"movieexample.com/pkg/utilities"

	otelruntimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
)

func Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return otel.Meter(instrumentationName, opts...)
}

func ObtainMetricCounter(name, desc string) metric.Int64Counter {
	counter, err := Meter("gotrue").Int64Counter(name, metric.WithDescription(desc))
	if err != nil {
		panic(err)
	}
	return counter
}

func enablePrometheusMetrics(ctx context.Context, mc *MetricsConfig) error {
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	otel.SetMeterProvider(provider)

	cleanupWaitGroup.Add(1)
	go func() {
		addr := net.JoinHostPort(mc.PrometheusListenHost, mc.PrometheusListenPort)
		baseContext, cancel := context.WithCancel(context.Background())

		server := &http.Server{
			Addr:    addr,
			Handler: promhttp.Handler(),
			BaseContext: func(net.Listener) context.Context {
				return baseContext
			},
			ReadHeaderTimeout: 2 * time.Second, // to mitigate a Slowloris attack
		}

		go func() {
			defer cleanupWaitGroup.Done()
			<-ctx.Done()

			cancel() // close baseContext

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				// logrus.WithError(err).Errorf("prometheus server (%s) failed to gracefully shut down", addr)
				log.Printf("prometheus server (%s) failed to gracefully shut down: %v", addr, err)
			}
		}()

		// logrus.Infof("prometheus server listening on %s", addr)
		log.Printf("prometheus server listening on %s", addr)

		if err := server.ListenAndServe(); err != nil {
			// logrus.WithError(err).Errorf("prometheus server (%s) shut down", addr)
			log.Printf("prometheus server (%s) shut down: %v", addr, err)
		} else {
			// logrus.Info("prometheus metric exporter shut down")
			log.Println("prometheus metric exporter shut down")
		}
	}()

	return nil
}

func enableOpenTelemetryMetrics(ctx context.Context, mc *MetricsConfig) error {
	switch mc.ExporterProtocol {
	case "grpc":
		metricExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return err
		}
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		)

		otel.SetMeterProvider(meterProvider)

		cleanupWaitGroup.Add(1)
		go func() {
			defer cleanupWaitGroup.Done()

			<-ctx.Done()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := metricExporter.Shutdown(shutdownCtx); err != nil {
				// logrus.WithError(err).Error("unable to gracefully shut down OpenTelemetry metric exporter")
				log.Printf("unable to gracefully shut down OpenTelemetry metric exporter: %v", err)
			} else {
				// logrus.Info("OpenTelemetry metric exporter shut down")
				log.Println("OpenTelemetry metric exporter shut down")
			}
		}()

	case "http/protobuf":
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return err
		}
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		)

		otel.SetMeterProvider(meterProvider)

		cleanupWaitGroup.Add(1)
		go func() {
			defer cleanupWaitGroup.Done()

			<-ctx.Done()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := metricExporter.Shutdown(shutdownCtx); err != nil {
				// logrus.WithError(err).Error("unable to gracefully shut down OpenTelemetry metric exporter")
				log.Printf("unable to gracefully shut down OpenTelemetry metric exporter: %v", err)
			} else {
				// logrus.Info("OpenTelemetry metric exporter shut down")
				log.Println("")
			}
		}()

	default: // http/json for example
		return fmt.Errorf("unsupported OpenTelemetry exporter protocol %q", mc.ExporterProtocol)
	}
	// logrus.Info("OpenTelemetry metrics exporter started")
	log.Println("OpenTelemetry metrics exporter started")
	return nil

}

var (
	metricsOnce *sync.Once = &sync.Once{}
)

func ConfigureMetrics(ctx context.Context, mc *MetricsConfig) error {
	if ctx == nil {
		panic("context must not be nil")
	}

	var err error

	metricsOnce.Do(func() {
		if mc.Enabled {
			switch mc.Exporter {
			case Prometheus:
				if err = enablePrometheusMetrics(ctx, mc); err != nil {
					// logrus.WithError(err).Error("unable to start prometheus metrics exporter")
					log.Printf("unable to start prometheus metrics exporter: %v", err)
					return
				}

			case OpenTelemetryMetrics:
				if err = enableOpenTelemetryMetrics(ctx, mc); err != nil {
					// logrus.WithError(err).Error("unable to start OTLP metrics exporter")
					log.Printf("unable to start OTLP metrics exporter: %v", err)

					return
				}
			}
		}

		if err := otelruntimemetrics.Start(otelruntimemetrics.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
			// logrus.WithError(err).Error("unable to start OpenTelemetry Go runtime metrics collection")
			log.Printf("unable to start OpenTelemetry Go runtime metrics collection: %v", err)

		} else {
			// logrus.Info("Go runtime metrics collection started")
			log.Println("Go runtime metrics collection started")
		}

		meter := otel.Meter("gotrue")
		_, err := meter.Int64ObservableGauge(
			"gotrue_running",
			metric.WithDescription("Whether GoTrue is running (always 1)"),
			metric.WithInt64Callback(func(_ context.Context, obsrv metric.Int64Observer) error {
				obsrv.Observe(int64(1))
				return nil
			}),
		)
		if err != nil {
			// logrus.WithError(err).Error("unable to get gotrue.gotrue_running gague metric")
			log.Printf("unable to get gotrue.gotrue_running gague metric: %v", err)
			return
		}
	})

	return err
}

var (
	cleanupWaitGroup sync.WaitGroup
)

// WaitForCleanup waits until all observability long-running goroutines shut
// down cleanly or until the provided context signals done.
func WaitForCleanup(ctx context.Context) {
	utilities.WaitForCleanup(ctx, &cleanupWaitGroup)
}

type MetricsExporter = string

const (
	Prometheus           MetricsExporter = "prometheus"
	OpenTelemetryMetrics MetricsExporter = "opentelemetry"
)

type MetricsConfig struct {
	Enabled bool

	Exporter MetricsExporter `default:"opentelemetry"`

	// ExporterProtocol is the OTEL_EXPORTER_OTLP_PROTOCOL env variable,
	// only available when exporter is opentelemetry. See:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md
	ExporterProtocol string `default:"http/protobuf" envconfig:"OTEL_EXPORTER_OTLP_PROTOCOL"`

	PrometheusListenHost string `default:"0.0.0.0" envconfig:"OTEL_EXPORTER_PROMETHEUS_HOST"`
	PrometheusListenPort string `default:"9100" envconfig:"OTEL_EXPORTER_PROMETHEUS_PORT"`
}

func (mc MetricsConfig) Validate() error {
	return nil
}
