package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
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
	config "movieexample.com/metadata/configs"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	"movieexample.com/metadata/internal/repository/memory"
	mysql "movieexample.com/metadata/internal/repository/sql"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/discovery/tracing"
)

const serviceName = "metadata"

func main() {
	// logger startup
	var env string
	flag.StringVar(&env, "env", "dev", "environment")
	flag.Parse()
	logger, _ := zap.NewProduction()

	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}()

	logger = logger.With(zap.String("service", serviceName))

	// setting up metadata config rom yaml file
	var cfg *config.Config
	var port int

	{

		f, err := os.Open("./metadata/configs/base.yaml")
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

	var tp *tracesdk.TracerProvider

	{
		// Tracing setup
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

	}

	{

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
				if err := consulRegistry.ReportHealthState(instanceID); err != nil {
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

	}

	var repo metadata.Repository

	{

		// setting up repository Env
		var err error
		if env == "dev" {
			repo = memory.New()
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
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}
}
