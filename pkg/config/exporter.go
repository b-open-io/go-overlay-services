package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/gookit/slog"
	"gopkg.in/yaml.v2"
)

// ToJSON writes the configuration to a JSON file
func (l *Load) ToJSON(filePath string) error {
	data, err := json.MarshalIndent(l.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	slog.Infof("JSON config exported to: %s", filePath)
	return nil
}

// ToYAML writes the configuration to a YAML file
func (l *Load) ToYAML(filePath string) error {
	data, err := yaml.Marshal(l.cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}
	slog.Infof("YAML config exported to: %s", filePath)
	return nil
}

// ToEnv writes the configuration to an environment file
func (l *Load) ToEnv(filePath string) error {
	data := map[string]any{}
	if err := mapstructure.Decode(l.cfg, &data); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	flattened := map[string]string{}
	flattenConfig("", flattened, data)

	var content string
	for key, value := range flattened {
		content += fmt.Sprintf("%s=%s\n", key, value)
	}

	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	slog.Infof("ENV config exported to: %s", filePath)
	return nil
}

func flattenConfig(prefix string, out map[string]string, v any) {
	switch t := v.(type) {
	case map[string]any:
		for key, value := range t {
			newPrefix := strings.ToUpper(strings.TrimPrefix(prefix+"_"+key, "_"))
			flattenConfig(newPrefix, out, value)
		}
	case []any:
		var strValues []string
		for _, val := range t {
			strValues = append(strValues, formatAny(val))
		}
		out[prefix] = strings.Join(strValues, ",")
	case string:
		out[prefix] = t
	case bool, int, float64:
		out[prefix] = formatAny(t)
	}
}

func formatAny(val any) string {
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(v, "\n", ""), "\r", ""), "\t", ""))
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
