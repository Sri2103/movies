package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	"movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/discovery/tracing"
)

const serviceName = "metadata"

func main() {
	// logger startup
	logger, _ := zap.NewProduction()

	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}()

	logger = logger.With(zap.String("service", serviceName))

	// setting up metadata config rom yaml file

	f, err := os.Open("./metadata/configs/base.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}

	var cfg *config.Config

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration:%w", zap.Error(err))
	}

	port := cfg.API.Port

	logger.Info("Starting the metadata service", zap.Int("port", port))

	logger.Info("Configuration loaded", zap.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setting up metrics
	// mp,err := metrics.SetUpMetrics(ctx, serviceName)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize metrics", zap.Error(err))
	// }
	// defer mp.Shutdown(ctx)

	// setting consul registry

	// setup copy tracing
	tp, err := tracing.SetUpTracing(ctx, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize tracing", zap.Error(err))
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	// setting up metrics -2
	// mp, err := metrics.InitMetrics(ctx, serviceName)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize metrics", zap.Error(err))
	// }
	// defer mp.Shutdown(ctx)
	// go metrics.StartMetricsEndPoint(cfg.Prometheus.MetricsPort)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	consulRegistry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := consulRegistry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}

	// reporting healthy state
	go func() {
		for {
			if err := consulRegistry.ReportHealthState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}

			time.Sleep(1 * time.Second)
		}
	}()

	defer func() {
		err := consulRegistry.DeRegister(ctx, instanceID, serviceName)
		if err != nil {
			logger.Error("Failed to deregister service", zap.Error(err))
		}
	}()

	// setting up repository
	repo := memory.New()
	ctrl := metadata.New(repo)
	h := grpchandler.New(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// setting up grpc server
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithPropagators(propagation.TraceContext{}),
			otelgrpc.WithTracerProvider(tp),
		)),
	)
	// grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	reflection.Register(srv)
	gen.RegisterMetadataServiceServer(srv, h)

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
