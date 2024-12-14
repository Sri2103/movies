package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	memoryDiscovery "movieexample.com/pkg/discovery/memory"
	"movieexample.com/pkg/discovery/tracing"
	config "movieexample.com/rating/internal/configs"
	"movieexample.com/rating/internal/controller/rating"
	grpcHandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/repository/memory"
)

const serviceName = "rating"
const (
	devEnv = "dev"
)

func main() {
	var logger *zap.Logger
	// get env from env

	configPathLocal := "./rating/internal/configs/base.yaml"
	env := os.Getenv("ENV")
	configpath := os.Getenv("CONFIG_PATH")

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

	var cfg *config.Config
	var httpPort int
	var grpcPort int

	{
		if configpath == "" {
			configpath = configPathLocal
		}

		f, err := os.Open(configpath)
		if err != nil {
			logger.Fatal("Failed to open config file", zap.Error(err))
		}
		if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
			logger.Fatal("Failed to decode config file", zap.Error(err))
		}
		httpPort = cfg.API.Port
		grpcPort = cfg.GRPC.Port

	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setting up registry
	{
		var registry discovery.Registry
		var err error
		if env == devEnv {
			registry = memoryDiscovery.NewRegistry()
		} else {
			registry, err = consul.NewRegistry("consul:8500")
			if err != nil {
				panic(err)
			}
		}

		instanceID := discovery.GenerateInstanceID(serviceName)

		if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", cfg.GRPC.Port)); err != nil {
			panic(err)
		}

		// reporting healthy state
		if env != devEnv {
			go func() {
				for {
					if err := registry.ReportHealthState(instanceID); err != nil {
						logger.Error("Failed to report healthy state", zap.Error(err))
					}

					time.Sleep(1 * time.Second)
				}
			}()
		}

		defer func() {
			err := registry.DeRegister(ctx, instanceID, serviceName)
			if err != nil {
				logger.Error("Failed to deregister service", zap.Error(err))
			}
		}()

	}

	var wg sync.WaitGroup

	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// setting up tracing
	var tp *tracesdk.TracerProvider
	{
		var err error

		tp, err = tracing.SetUpTracing(ctx, serviceName)
		if err != nil {
			logger.Fatal("Failed to create tracer", zap.Error(err))
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

	// grpc server
	var grpcServer *grpc.Server

	{
		repo := memory.New()

		controller := rating.NewController(repo, nil)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			logger.Fatal("Failed to listen", zap.Error(err))
		}
		grpcServer = grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler(
				otelgrpc.WithPropagators(propagation.TraceContext{}),
				otelgrpc.WithTracerProvider(tp),
			)),
		)

		reflection.Register(grpcServer)
		gen.RegisterRatingServiceServer(grpcServer, grpcHandler.New(controller))
		logger.Info("Starting  grpc rating service on port", zap.Int("port", grpcPort))

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := grpcServer.Serve(lis); err != nil {
				logger.Fatal("Failed to serve", zap.Error(err))
			}
		}()

	}

	// http server to readiness and liveness probe
	var serverHTTP *http.Server
	{
		// var err error

		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		serverHTTP = &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: mux,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("Starting http live and ready server", zap.Int("port", httpPort))
			if err := serverHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start http server", zap.Error(err))
			}
		}()
	}

	// cleanup

	<-stop
	logger.Info("Shutting down")

	shutCtx, shutCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutCancel()

	if err := serverHTTP.Shutdown(shutCtx); err != nil {
		logger.Error("Failed to shutdown http server", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped")
	}

	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped")

	wg.Wait()
	logger.Info("Shutdown complete")
	os.Exit(0)
}
