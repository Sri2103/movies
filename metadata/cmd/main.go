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

	"github.com/spf13/viper"
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
	memoryRepo "movieexample.com/metadata/internal/repository/memory"
	mysql "movieexample.com/metadata/internal/repository/sql"
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
	configPathLocal := "./metadata/configs/base.yaml"
	env := os.Getenv("ENV")
	configpath := os.Getenv("CONFIG_PATH")
	if env == "" {
		env = devEnv
	}
	if configpath == "" {
		configpath = configPathLocal
	}
	if env == devEnv {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	viperConfig := viper.New()
	logger.Info("env", zap.String("env", env))
	logger.Info("configpath", zap.String("configpath", configpath))
	viperConfig.SetConfigFile(configpath)
	viperConfig.SetConfigType("yaml")
	if err := viperConfig.ReadInConfig(); err != nil {
		logger.Fatal("Failed to read configuration", zap.Error(err))
	}

	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	logger = logger.With(zap.String("service", serviceName))

	// setting up metadata config rom yaml file
	var cfg *config.Config
	var port int

	{

		f, err := os.Open(configpath)
		if err != nil {
			logger.Fatal("Failed to open configuration", zap.Error(err))
		}

		if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
			logger.Fatal("Failed to parse configuration:%w", zap.Error(err))
		}

		port = cfg.API.Port

		logger.Info("Starting the metadata service", zap.Int("port", port))

	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	var repo metadata.Repository

	var wg sync.WaitGroup
	var srv *grpc.Server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	{
		tp, err := tracing.SetUpTracing(ctx, serviceName)
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
		if env == devEnv {
			repo = memoryRepo.New()
		} else {
			repo, err = mysql.New()
			logger.Info("Connected to mysql")
			if err != nil {
				logger.Fatal("Failed to initialize mysql", zap.Error(err))
			}
		}

		// setting up repository
		ctrl := metadata.New(repo)
		h := grpchandler.New(ctrl)

		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", cfg.GRPC.Port))
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
		log.Println("HTTP port binding here:", cfg.API.Port)

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
