package config

type Config struct {
	Consul Consul `yaml:"consul"`
	Grpc   Grpc   `yaml:"grpc"`
	API    API    `yaml:"http"`
}

type Consul struct {
	Host string `yaml:"host"`
}

type Grpc struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Jaeger struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type API struct {
	Port int `yaml:"port"`
}
