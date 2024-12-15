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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"movieexample.com/gen"
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	memoryRepo "movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/metadata/internal/repository/postgres"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	memoryDiscovery "movieexample.com/pkg/discovery/memory"
	"movieexample.com/pkg/discovery/tracing"
)

const (
	serviceName = "metadata"
	devEnv      = "dev"
	prodEnv     = "prod"
)

func main() {
	var logger *zap.Logger

	cfg, err := config.SetUpConfig()
	if err != nil {
		panic(err)
	}
	env := os.Getenv("ENV")
	fmt.Println(env, "Env from Environment")
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

	// setting up metadata config rom yaml file

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
					if err := registry.ReportHealthState(instanceID,serviceName); err != nil {
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

	var repo metadata.Repository

	var wg sync.WaitGroup
	var srv *grpc.Server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	{
		tp, err := tracing.SetUpTracing(ctx, serviceName, cfg.Jaeger.URL)
		if err != nil {
			logger.Fatal("Failed to initialize tracing", zap.Error(err))
		}

		defer func() {
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown tracer provider", zap.Error(err))
			}
		}()

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})

		// setting up repository Env
		if env == string(devEnv) {
			repo = memoryRepo.New()
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

		// setting up repository
		ctrl := metadata.New(repo)
		h := grpchandler.New(ctrl)

		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%v", cfg.Host, cfg.GRPC.Port))
		if err != nil {
			logger.Fatal("Failed to listen", zap.Error(err))
		}

		srv = grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler(
				otelgrpc.WithPropagators(propagation.TraceContext{}),
				otelgrpc.WithTracerProvider(tp),
			)),
		)
		// grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
		reflection.Register(srv)
		gen.RegisterMetadataServiceServer(srv, h)

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("Starting gRPC server", zap.Int("port", cfg.GRPC.Port))
			if err := srv.Serve(lis); err != nil {
				logger.Fatal("Failed to serve", zap.Error(err))
				return
			}
		}()

	}

	httpServ := http.NewServeMux()
	httpServ.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	httpServ.HandleFunc("/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.API.Port),
		Handler: httpServ,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("Starting health check server", zap.Int("port", cfg.API.Port))
		logger.Info("HTTP port binding here:", zap.Int("http_port", cfg.API.Port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Info("HTTP server failed: ", zap.Error(err))
			return
		}
	}()

	<-stop
	logger.Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	} else {
		logger.Info("HTTP server stopped")
	}

	// graceful shutdown of gRPC server
	srv.GracefulStop()
	logger.Info("Grpc Server stopped")
	wg.Wait()
	logger.Info("All servers stopped")
}
