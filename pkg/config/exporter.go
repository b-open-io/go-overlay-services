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

func (l *Load) ToJSON(filePath string) error {
	data, err := json.MarshalIndent(l.cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	slog.Infof("JSON config exported to: %s", filePath)
	return nil
}

func (l *Load) ToYAML(filePath string) error {
	data, err := yaml.Marshal(l.cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	slog.Infof("YAML config exported to: %s", filePath)
	return nil
}

func (l *Load) ToEnv(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]any{}
	if err := mapstructure.Decode(l.cfg, &data); err != nil {
		return err
	}

	flattened := map[string]string{}
	flattenConfig("", flattened, data)

	for key, value := range flattened {
		_, err := file.WriteString(key + "=" + value + "\n")
		if err != nil {
			return err
		}
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
