package configs

// config defines the configuration for the application.
// It includes settings for the API and Jaeger.
type Config struct {
	// API holds the configuration for the API.
	API apiConfig `yaml:"api"`
	// Jaeger holds the configuration for Jaeger.
	Jaeger jaegerConfig `yaml:"jaeger"`
}

// apiConfig defines the configuration for the API.
// It includes the port the API should listen on.
type apiConfig struct {
	// Port is the port the API should listen on.
	Port int `yaml:"port"`
}

// jaegerConfig defines the configuration for Jaeger.
// It includes the URL of the Jaeger instance.
type jaegerConfig struct {
	// URL is the URL of the Jaeger instance.
	URL string `yaml:"url"`
}
