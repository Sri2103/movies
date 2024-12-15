package config

type Config struct {
	API        *APIConfig        `yaml:"http"`
	Jaeger     *JaegerConfig     `yaml:"jaeger"`
	Prometheus *PrometheusConfig `yaml:"prometheus"`
	Consul     *ConsulConfig     `yaml:"consul"`
	GRPC       *GRPCConfig       `yaml:"grpc"`
	Host       string            `yaml:"host"`
	Postgres   *PostgresConfig   `yaml:"mysql"`
}

type APIConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type GRPCConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type JaegerConfig struct {
	URL string `yaml:"url"`
}

type PrometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}

type ConsulConfig struct {
	Address string `yaml:"address"`
}
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SslMode  string `yaml:"sslmode"`
}
