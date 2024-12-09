package config

type Config struct {
	API        APIConfig        `yaml:"api"`
	Jaeger     JaegerConfig     `yaml:"jaeger"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

type APIConfig struct {
	Port int `yaml:"port"`
}

type JaegerConfig struct {
	URL string `yaml:"url"`
}

type PrometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}
