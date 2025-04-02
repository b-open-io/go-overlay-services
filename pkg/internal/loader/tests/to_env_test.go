package loader_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/4chain-ag/go-overlay-services/pkg/internal/loader"
	"github.com/stretchr/testify/require"
)

func TestToEnvFile(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	configFilePath := fmt.Sprintf("%s/exported_config.env", tmpDir)
	cfg := Defaults()

	// When
	err := loader.ToEnvFile(cfg, configFilePath)

	// Then
	require.NoError(t, err)

	data, err := os.ReadFile(configFilePath)
	require.NoError(t, err)

	content := string(data)

	require.Contains(t, content, `A="default_hello"`)
	require.Contains(t, content, `B_WITH_LONG_NAME="1"`)
	require.Contains(t, content, `C_SUB_CONFIG_D_NESTED_FIELD="default_world"`)
}
