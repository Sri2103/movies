package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/movie/internal/configs"
	"movieexample.com/movie/internal/controller/movie"
	metadatagateway "movieexample.com/movie/internal/gateway/metadata/http"
	ratinggateway "movieexample.com/movie/internal/gateway/rating/http"
	grpcHandler "movieexample.com/movie/internal/handler/grpc"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/discovery/tracing"
)

const ServiceName = "movie"

func main() {
	logger, _ := zap.NewProduction()

	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}()

	f, err := os.Open("./movie/internal/configs/base.yaml")
	if err != nil {
		logger.Fatal("Failed to open config file", zap.Error(err))
	}

	var cfg configs.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to decode config file", zap.Error(err))
	}

	port := cfg.API.Port

	logger.Info("Starting the movie service on port", zap.Int("port", port))

	ctx := context.Background()

	// setting up tracer
	grpcTracer, err := tracing.NewGrpcTracerProvider(cfg.Jaeger.URL, ServiceName)
	if err != nil {
		logger.Fatal("Failed to create tracer provider", zap.Error(err))
	}

	otel.SetTracerProvider(grpcTracer)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		logger.Fatal("Failed to create registry", zap.Error(err))
	}

	instanceID := discovery.GenerateInstanceID(ServiceName)
	if err := registry.Register(ctx, instanceID, ServiceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		logger.Fatal("Failed to register service", zap.Error(err))
	}

	go func() {
		for {
			if err := registry.ReportHealthState(instanceID); err != nil {
				log.Printf("Failed to report healthy state: %s", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	defer func() {
		if err := registry.ReportHealthState(instanceID); err != nil {
			log.Printf("Failed to report unhealthy state: %s", err)
		}
	}()

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)

	ctrl := movie.New(ratingGateway, metadataGateway)
	startGrpcServer(port, ctrl)
}

func startGrpcServer(port int, controller *movie.Controller) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	gen.RegisterMovieServiceServer(grpcServer, grpcHandler.New(controller))
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to drive server: %v", err)
	}
}
