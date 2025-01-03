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
	"movieexample.com/gen"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	memoryDiscovery "movieexample.com/pkg/discovery/memory"
	"movieexample.com/pkg/discovery/tracing"
	config "movieexample.com/rating/configs"
	"movieexample.com/rating/internal/controller/rating"
	grpcHandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/repository/memory"
	"movieexample.com/rating/internal/repository/postgres"
)

const serviceName = "rating"
const (
	devEnv = "dev"
)

func main() {
	var logger *zap.Logger
	// get env from env
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

	logger = logger.With(zap.String("service", serviceName))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setting up registry
	{
		var registry discovery.Registry
		var err error
		if env == devEnv {
			registry = memoryDiscovery.NewRegistry()
		} else {
			registry, err = consul.NewRegistry(cfg.Consul.Address)
			if err != nil {
				panic(err)
			}
		}

		instanceID := discovery.GenerateInstanceID(serviceName)

		if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPC.Port)); err != nil {
			panic(err)
		}

		// reporting healthy state
		if env != devEnv {
			go func() {
				for {
					if err := registry.ReportHealthState(instanceID, serviceName); err != nil {
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

		tp, err = tracing.SetUpTracing(ctx, serviceName, cfg.Jaeger.URL)
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
	var repo rating.Repository

	{
		if env == string(devEnv) {
			repo = memory.New()
			logger.Info("Connected to memory")
			logger.Info("Memory repo connected at env: ", zap.String("env", env))
		} else {
			repo, err = postgres.ConnectSQL(ctx, cfg)
			if err != nil {
				logger.Fatal("Failed to initialize mysql", zap.Error(err))
			}
			logger.Info("Connected to DB")
			defer postgres.CloseDB(repo)
		}

		controller := rating.NewController(repo, nil)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
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
		logger.Info("Starting  grpc rating service on port", zap.Int("port", cfg.GRPC.Port))

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
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		serverHTTP = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.API.Port),
			Handler: mux,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("Starting http live and ready server", zap.Int("port", cfg.API.Port))
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
