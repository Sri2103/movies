package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"movieexample.com/gen"
	config "movieexample.com/movie/configs"
	"movieexample.com/movie/internal/controller/movie"
	metadatagateway "movieexample.com/movie/internal/gateway/metadata/http"
	ratinggateway "movieexample.com/movie/internal/gateway/rating/http"
	grpcHandler "movieexample.com/movie/internal/handler/grpc"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/discovery/memory"
	"movieexample.com/pkg/discovery/tracing"
)

const ServiceName = "movie"

const (
	devEnv = "dev"
)

func main() {
	var logger *zap.Logger
	cfg, err := config.SetUpConfig()
	if err != nil {
		panic(err)
	}

	env := os.Getenv("ENV")

	if env == "" {
		env = devEnv
	}

	if env == devEnv {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setting up tracer
	var tp *tracesdk.TracerProvider
	{
		var err error
		tp, err = tracing.SetUpTracing(ctx, ServiceName, cfg.Jaeger.URL)
		if err != nil {
			logger.Fatal("Failed to create tracer provider", zap.Error(err))
		}
		defer func() {
			err := tp.Shutdown(ctx)
			if err != nil {
				logger.Error("Failed to shutdown tracer provider", zap.Error(err))
			}
		}()

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
	}

	var registry discovery.Registry
	{

		var err error

		if env == devEnv {
			registry = memory.NewRegistry()
		} else {
			registry, err = consul.NewRegistry(cfg.Consul.Address)
			if err != nil {
				logger.Fatal("error connecting to consul", zap.Error(err))
			}
		}

		instanceID := discovery.GenerateInstanceID(ServiceName)
		if err := registry.Register(ctx, instanceID, ServiceName, fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPC.Port)); err != nil {
			logger.Fatal("Failed to register service", zap.Error(err))
		}

		if env != devEnv {
			go func() {
				for {
					if err := registry.ReportHealthState(instanceID, ServiceName); err != nil {
						logger.Error("Failed to report healthy state", zap.Error(err))
					}

					time.Sleep(1 * time.Second)
				}
			}()
		}
		defer func() {
			err := registry.DeRegister(ctx, instanceID, ServiceName)
			if err != nil {
				logger.Error("Failed to deregister service", zap.Error(err))
			}
		}()

	}

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)

	var wg sync.WaitGroup
	var grpcServer *grpc.Server

	{
		controller := movie.New(ratingGateway, metadataGateway)
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", cfg.GRPC.Port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer = grpc.NewServer()
		reflection.Register(grpcServer)
		gen.RegisterMovieServiceServer(grpcServer, grpcHandler.New(controller))
		log.Printf("server listening at %v", lis.Addr())
		wg.Add(1)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("failed to drive server: %v", err)
			}
		}()

	}

	var httpServer *http.Server
	{
		server := http.NewServeMux()
		server.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.API.Port),
			Handler: server,
		}
		wg.Add(1)
		go func() {
			wg.Done()
			err := httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Sugar().Fatalf("Server stopped unfortunately", zap.Error(err))
			}
		}()
	}

	<-stop
	logger.Warn("Stopping the servers")
	shutctx, shutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutCancel()
	if err := httpServer.Shutdown(shutctx); err != nil {
		logger.Error("Failed to stop http server", zap.Error(err))
	}
	grpcServer.GracefulStop()
	logger.Warn("GRPC server stopped")
	wg.Wait()
	logger.Warn("All servers cleared")
	os.Exit(0)
}
