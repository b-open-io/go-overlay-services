package config

// Loader defines methods for loading configuration.
type Loader interface {
	Load() (ServerConfig, error)
	SetConfigFilePath(path string) error
}

// Exporter defines methods for exporting configuration.
type Exporter interface {
	ToYAML(filePath string) error
	ToJSON(filePath string) error
	ToEnv(filePath string) error
}
