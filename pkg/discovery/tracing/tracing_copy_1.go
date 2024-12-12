package tracing

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"movieexample.com/pkg/utilities"
)

func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

func openTelemetryResource() *sdkresource.Resource {
	environmentResource := sdkresource.Environment()
	gotrueResource := sdkresource.NewSchemaless(attribute.String("gotrue.version", "v1"))

	mergedResource, err := sdkresource.Merge(environmentResource, gotrueResource)
	if err != nil {
		// logrus.WithError(err).Error("unable to merge OpenTelemetry environment and gotrue resources")
		log.Println("unable to merge OpenTelemetry environment and gotrue resources")

		return environmentResource
	}

	return mergedResource
}

func enableOpenTelemetryTracing(ctx context.Context, tc *Config) error {
	var (
		err           error
		traceExporter *otlptrace.Exporter
	)

	switch tc.ExporterProtocol {
	case "grpc":
		traceExporter, err = otlptracegrpc.New(ctx)
		if err != nil {
			return err
		}

	case "http/protobuf":
		traceExporter, err = otlptracehttp.New(ctx)
		if err != nil {
			return err
		}

	default: // http/json for example
		return fmt.Errorf("unsupported OpenTelemetry exporter protocol %q", tc.ExporterProtocol)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(openTelemetryResource()),
	)

	otel.SetTracerProvider(traceProvider)

	// Register the W3C trace context and baggage propagators so data is
	// propagated across services/processes
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	cleanupWaitGroup.Add(1)
	go func() {
		defer cleanupWaitGroup.Done()

		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := traceExporter.Shutdown(shutdownCtx); err != nil {
			// logrus.WithError(err).Error("unable to shutdown OpenTelemetry trace exporter")
			panic(err)
		}

		if err := traceProvider.Shutdown(shutdownCtx); err != nil {
			// logrus.WithError(err).Error("unable to shutdown OpenTelemetry trace provider")
			panic(err)
		}
	}()

	// logrus.Info("OpenTelemetry trace exporter started")
	log.Println("OpenTelemetry trace exporter started")

	return nil
}

var tracingOnce sync.Once

// ConfigureTracing sets up global tracing configuration for OpenTracing /
// OpenTelemetry. The context should be the global context. Cancelling this
// context will cancel tracing collection.
func ConfigureTracing(ctx context.Context, tc *Config) error {
	if ctx == nil {
		panic("context must not be nil")
	}

	var err error

	tracingOnce.Do(func() {
		if tc.Enabled {
			if tc.Exporter == OpenTelemetryTracing {
				if err = enableOpenTelemetryTracing(ctx, tc); err != nil {
					// logrus.WithError(err).Error("unable to start OTLP trace exporter")
					log.Println("unable to start OTLP trace exporter")
				}
			}
		}
	})

	return err
}

type Exporter = string

const (
	OpenTelemetryTracing Exporter = "opentelemetry"
)

type Config struct {
	Enabled  bool
	Exporter Exporter `default:"opentelemetry"`

	// ExporterProtocol is the OTEL_EXPORTER_OTLP_PROTOCOL env variable,
	// only available when exporter is opentelemetry. See:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md
	ExporterProtocol string `default:"http/protobuf" envconfig:"OTEL_EXPORTER_OTLP_PROTOCOL"`

	// Host is the host of the OpenTracing collector.
	Host string

	// Port is the port of the OpenTracing collector.
	Port string

	// ServiceName is the service name to use with OpenTracing.
	ServiceName string `default:"gotrue" split_words:"true"`

	// Tags are the tags to associate with OpenTracing.
	Tags map[string]string
}

func (tc *Config) Validate() error {
	return nil
}

var cleanupWaitGroup sync.WaitGroup

// WaitForCleanup waits until all observability long-running goroutines shut
// down cleanly or until the provided context signals done.
func WaitForCleanup(ctx context.Context) {
	utilities.WaitForCleanup(ctx, &cleanupWaitGroup)
}
