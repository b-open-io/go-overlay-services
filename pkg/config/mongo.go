package config

import "fmt"

// MongoDB is the configuration struct for MongoDB connections.
type MongoDB struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	AuthDB   string `mapstructure:"auth_db"` // Optional: for auth source
}

// DefaultMongo provides the default MongoDB configuration.
func DefaultMongo() MongoDB {
	return MongoDB{
		URI:      "mongodb://localhost:27017",
		Database: "overlay",
		Username: "",
		Password: "",
		AuthDB:   "admin",
	}
}

// validate validates the MongoDB configuration.
func (cfg *MongoDB) validate() error {
	if cfg.URI == "" {
		return fmt.Errorf("mongodb URI must not be empty")
	}
	if cfg.Database == "" {
		return fmt.Errorf("mongodb database must not be empty")
	}
	return nil
}

func (cfg *MongoDB) ValidateCreds() bool {
	return cfg.Username != "" && cfg.Password != ""
}
