package config

// SyncOption represents the sync option
type SyncOption struct {
	Keys []string `mapstructure:"keys"`
}

// SyncConfiguration represents the sync configuration
type SyncConfiguration map[string]SyncOption

// EngineConfig represents the engine configuration
type EngineConfig struct {
	ChainTracker            string            `mapstructure:"chain_tracker"`
	SyncConfiguration       SyncConfiguration `mapstructure:"sync_configuration"`
	LogTime                 bool              `mapstructure:"log_time"`
	LogPrefix               string            `mapstructure:"log_prefix"`
	ThrowOnBroadcastFailure bool              `mapstructure:"throw_on_broadcast_failure"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

// MongoConfig represents the mongo configuration
type MongoConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	ConnectionString string `mapstructure:"connection_string"`
}

// MigrationConfig represents the migration configuration
type MigrationConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	MigrationsDir string `mapstructure:"migrations_dir"`
}

// LoggerConfig represents the logger configuration
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	PrettyPrint bool   `mapstructure:"pretty_print"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Name                  string          `mapstructure:"name"`
	PrivateKey            string          `mapstructure:"private_key"`
	HostingURL            string          `mapstructure:"hosting_url"`
	Address               string          `mapstructure:"address"`
	Port                  int             `mapstructure:"port"`
	Network               string          `mapstructure:"network"`
	EnableGASPSync        bool            `mapstructure:"enable_gasp_sync"`
	ArcApiKey             string          `mapstructure:"arc_api_key"`
	VerboseRequestLogging bool            `mapstructure:"verbose_request_logging"`
	EngineConfig          EngineConfig    `mapstructure:"engine_config"`
	DatabaseConfig        DatabaseConfig  `mapstructure:"database_config"`
	MongoConfig           MongoConfig     `mapstructure:"mongo_config"`
	MigrationConfig       MigrationConfig `mapstructure:"migration_config"`
	LoggerConfig          LoggerConfig    `mapstructure:"logger_config"`
}

// DefaultConfig returns the default server configuration
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Address: "localhost",
		Port:    3000,
		Network: "main",
		LoggerConfig: LoggerConfig{
			Level:       "debug",
			Format:      "json",
			PrettyPrint: true,
		},
		DatabaseConfig: DatabaseConfig{
			Enabled: false,
			URL:     "",
		},
		MongoConfig: MongoConfig{
			Enabled:          false,
			ConnectionString: "",
		},
		MigrationConfig: MigrationConfig{
			Enabled:       false,
			MigrationsDir: "./migrations",
		},
	}
}
