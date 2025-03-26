package config_test

import (
	"os"
	"path/filepath"
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
	require.Equal(t, "localhost", cfg.Addr)
	require.Equal(t, 3000, cfg.Port)
	require.Equal(t, "Overlay API v0.0.0", cfg.AppName)
	require.Equal(t, "Overlay API", cfg.ServerHeader)
	require.Equal(t, "admin-token-default", cfg.AdminBearerToken)
}

func TestLoad_ShouldOverrideDefaults_WhenConfigFileProvidesValues(t *testing.T) {
	// Given
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	content := `
addr: 127.0.0.1
port: 9999
appname: CustomApp
serverheader: CustomHeader
adminbearertoken: secret-token
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0600))

	loader := config.NewLoader("OVERLAY")
	require.NoError(t, loader.SetConfigFilePath(tmpFile))

	// When
	cfg, err := loader.Load()

	// Then
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", cfg.Addr)
	require.Equal(t, 9999, cfg.Port)
	require.Equal(t, "CustomApp", cfg.AppName)
	require.Equal(t, "CustomHeader", cfg.ServerHeader)
	require.Equal(t, "secret-token", cfg.AdminBearerToken)
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
	require.Contains(t, string(data), "addr: localhost")
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
	require.Contains(t, string(data), `"Addr": "localhost"`)
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
	require.Contains(t, content, "ADDR=localhost")
	require.Contains(t, content, "PORT=3000")
	require.Contains(t, content, "APP_NAME=Overlay API v0.0.0")
}

func TestExportToEnv_ShouldFail_WhenFilePathIsInvalid(t *testing.T) {
	// Given: a loaded config
	loader := config.NewLoader("OVERLAY")
	_, _ = loader.Load()

	// When: exporting to an invalid file path
	err := loader.ToEnv("/invalid/!!/envfile")

	// Then: an error should be returned
	require.Error(t, err)
}
