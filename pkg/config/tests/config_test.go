package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_ShouldApplyAllDefaults_WhenNoConfigFileExists(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")

	// When
	cfg, err := loader.Load()

	// Then
	require.NoError(t, err)
	require.Equal(t, "localhost", cfg.Address)
	require.Equal(t, 3000, cfg.Port)
	require.Equal(t, "main", cfg.Network)
	require.Equal(t, "debug", cfg.LoggerConfig.Level)
	require.False(t, cfg.DatabaseConfig.Enabled)
	require.False(t, cfg.MongoConfig.Enabled)
	require.False(t, cfg.MigrationConfig.Enabled)
	require.Equal(t, "./migrations", cfg.MigrationConfig.MigrationsDir)
}

func TestLoad_ShouldApplyDefaultConfig_WhenNoConfigFileExists(t *testing.T) {
	// Given: a new config loader with no config file
	loader := config.NewLoader("OVERLAY")

	// When: loading the configuration
	cfg, err := loader.Load()

	// Then: all default values should be correctly applied
	require.NoError(t, err)

	// Top-level server config
	require.Equal(t, "", cfg.Name)
	require.Equal(t, "", cfg.PrivateKey)
	require.Equal(t, "", cfg.HostingURL)
	require.Equal(t, "localhost", cfg.Address)
	require.Equal(t, 3000, cfg.Port)
	require.Equal(t, "main", cfg.Network)
	require.False(t, cfg.EnableGASPSync)
	require.Equal(t, "", cfg.ArcApiKey)
	require.False(t, cfg.VerboseRequestLogging)

	// Logger config
	require.Equal(t, "debug", cfg.LoggerConfig.Level)
	require.Equal(t, "json", cfg.LoggerConfig.Format)
	require.True(t, cfg.LoggerConfig.PrettyPrint)

	// Engine config
	require.Equal(t, "", cfg.EngineConfig.ChainTracker)
	require.False(t, cfg.EngineConfig.ThrowOnBroadcastFailure)
	require.Equal(t, "", cfg.EngineConfig.LogPrefix)
	require.False(t, cfg.EngineConfig.LogTime)
	require.Empty(t, cfg.EngineConfig.SyncConfiguration)

	// Database config
	require.False(t, cfg.DatabaseConfig.Enabled)
	require.Equal(t, "", cfg.DatabaseConfig.URL)

	// Mongo config
	require.False(t, cfg.MongoConfig.Enabled)
	require.Equal(t, "", cfg.MongoConfig.ConnectionString)

	// Migration config
	require.False(t, cfg.MigrationConfig.Enabled)
	require.Equal(t, "./migrations", cfg.MigrationConfig.MigrationsDir)
}

func TestSetConfigFilePath_ShouldReturnError_WhenUnsupportedExtension(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")

	// When
	err := loader.SetConfigFilePath("config.txt")

	// Then
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported config file extension")
}

func TestExportToYAML_ShouldWriteFile_WhenConfigIsValid(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")
	_, err := loader.Load()
	require.NoError(t, err)

	// When
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	err = loader.ToYAML(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(data), "address: localhost")
}

func TestExportToJSON_ShouldWriteFile_WhenConfigIsValid(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")
	_, err := loader.Load()
	require.NoError(t, err)

	// When
	tmpFile := filepath.Join(t.TempDir(), "config.json")
	err = loader.ToJSON(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(data), `"Address": "localhost"`)
}

func TestExportToEnv_ShouldWriteFlatEnvFile_WhenConfigIsValid(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")
	_, err := loader.Load()
	require.NoError(t, err)

	// When
	tmpFile := filepath.Join(t.TempDir(), ".env")
	err = loader.ToEnv(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	content := string(data)
	require.Contains(t, content, "ADDRESS=localhost")
	require.Contains(t, content, "PORT=3000")
	require.Contains(t, content, "LOGGER_CONFIG_LEVEL=debug")
}

func TestExportToEnv_ShouldFail_WhenFilePathIsInvalid(t *testing.T) {
	// Given
	loader := config.NewLoader("OVERLAY")
	_, _ = loader.Load()

	// When
	err := loader.ToEnv("/invalid/!!/envfile")

	// Then
	require.Error(t, err)
}

func TestToYAML_ShouldExportConfigToYAMLFile_WhenConfigIsLoaded(t *testing.T) {
	// Given
	loader := config.NewLoader("TEST")
	_, _ = loader.Load()

	// When
	tmpFile := filepath.Join(t.TempDir(), "test_config.yaml")
	err := loader.ToYAML(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(data), "port: 3000")
}

func TestToJSON_ShouldExportConfigToJSONFile_WhenConfigIsLoaded(t *testing.T) {
	// Given
	loader := config.NewLoader("TEST")
	_, _ = loader.Load()

	// When
	tmpFile := filepath.Join(t.TempDir(), "test_config.json")
	err := loader.ToJSON(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(data), `"Port": 3000`)
}

func TestToEnv_ShouldFlattenNestedStructsIntoENVKeys_WhenExporting(t *testing.T) {
	// Given
	loader := config.NewLoader("TEST")
	_, _ = loader.Load()

	// When
	tmpFile := filepath.Join(t.TempDir(), "flatten.env")
	err := loader.ToEnv(tmpFile)

	// Then
	require.NoError(t, err)
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	lines := strings.Split(string(data), "\n")
	require.GreaterOrEqual(t, len(lines), 5)
	require.Contains(t, string(data), "LOGGER_CONFIG_LEVEL=debug")
}
