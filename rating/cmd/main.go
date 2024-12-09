package main

import (
	"context"
	"fmt"
	"log"
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
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/discovery/tracing"
	"movieexample.com/rating/internal/configs"
	"movieexample.com/rating/internal/controller/rating"
	grpcHandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/repository/memory"
)

const ServiceName = "rating"

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}()

	f, err := os.Open("./rating/internal/configs/base.yaml")
	if err != nil {
		logger.Fatal("Failed to open config file", zap.Error(err))
	}
	var cfg configs.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to decode config file", zap.Error(err))
	}
	port := cfg.API.Port

	logger.Info("Starting the metadata service on port %d", zap.Int("port", port))

	ctx := context.Background()

	grpcTracer, err := tracing.NewGrpcTracerProvider(cfg.Jaeger.URL, ServiceName)
	if err != nil {
		logger.Fatal("Failed to create tracer", zap.Error(err))
	}
	defer func() {
		err := grpcTracer.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	otel.SetTracerProvider(grpcTracer)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// setting up registry

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		logger.Fatal("Failed to create registry", zap.Error(err))
	}

	instanceID := discovery.GenerateInstanceID(ServiceName)
	// if err := registry.Register(ctx, instanceID, ServiceName, fmt.Sprintf("localhost:%d", port)); err != nil {
	// 	logger.Fatal("Failed to register service", zap.Error(err))
	// }

	// go func() {
	// 	for {
	// 		if err := registry.ReportHealthState(instanceID, ServiceName); err != nil {
	// 			log.Printf("Failed to report healthy state: %s", err)
	// 		}
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()
	if err := registerConsul(ctx, registry, instanceID, ServiceName, port); err != nil {
		logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer func() {
		err := registry.DeRegister(ctx, instanceID, ServiceName)
		if err != nil {
			logger.Fatal("Failed to deregister service", zap.Error(err))
		}
	}()

	repo := memory.New()

	controller := rating.NewController(repo, nil)

	startGrpcServer(logger, port, controller)
}

func startGrpcServer(logger *zap.Logger, port int, controller *rating.Controller) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
	)
	reflection.Register(grpcServer)
	gen.RegisterRatingServiceServer(grpcServer, grpcHandler.New(controller))
	logger.Info("Starting rating service on port %d", zap.Int("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}

func registerConsul(ctx context.Context, registry discovery.Registry, instanceID string, serviceName string, port int) error {
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		return err
	}
	go func() {
		for {
			if err := registry.ReportHealthState(instanceID, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %s", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}
