package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func SetUpConfig() (*Config, error) {
	viperConfig := viper.New()
	err := godotenv.Load(".env.metadata")
	if err != nil {
		log.Println("Error loading .env file")
	}
	viperConfig.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{}

	jaegerHost := viperConfig.GetString("JAEGER_AGENT_HOST")
	jaegerPort := viperConfig.GetInt("JAEGER_AGENT_PORT")
	consulHost := viperConfig.GetString("CONSUL_HOST")
	consulPort := viperConfig.GetInt("CONSUL_PORT")
	httpPort := viperConfig.GetInt("HTTP_PORT")
	host := viperConfig.GetString("HOST")
	grpcPort := viperConfig.GetInt("GRPC_PORT")
	consulAddress := fmt.Sprintf("%s:%d", consulHost, consulPort)
	jaegerURL := fmt.Sprintf("%s:%d", jaegerHost, jaegerPort)
	postgresHost := viperConfig.GetString("POSTGRES_HOST")
	postgresPort := viperConfig.GetInt("POSTGRES_PORT")
	postgresUser := viperConfig.GetString("POSTGRES_USER")
	postgresPassword := viperConfig.GetString("POSTGRES_PASSWORD")
	postgresDatabase := viperConfig.GetString("POSTGRES_DATABASE")
	postgresSslMode := viperConfig.GetString("POSTGRES_SSL_MODE")

	cfg.API = &APIConfig{
		Host: host,
		Port: httpPort,
	}
	cfg.Consul = &ConsulConfig{
		Address: consulAddress,
	}
	cfg.Jaeger = &JaegerConfig{
		URL: jaegerURL,
	}

	cfg.GRPC = &GRPCConfig{
		Port: grpcPort,
	}
	cfg.Host = host
	cfg.Postgres = &PostgresConfig{
		Host:     postgresHost,
		Port:     postgresPort,
		Username: postgresUser,
		Password: postgresPassword,
		Database: postgresDatabase,
		SslMode:  postgresSslMode,
	}

	return cfg, nil
}
