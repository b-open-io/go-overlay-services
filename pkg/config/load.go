package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/gookit/slog"
	"github.com/spf13/viper"
)

// DefaultConfigFilePath is the default path to the configuration file
const DefaultConfigFilePath = "config.yaml"

// Load implements the Loader and Exporter interfaces
type Load struct {
	cfg            *Config
	envPrefix      string
	configFilePath string
	configFileExt  string
	viper          *viper.Viper
}

// NewLoader creates a new configuration loader
func NewLoader(envPrefix string) *Load {
	return &Load{
		cfg:            NewConfig(),
		envPrefix:      envPrefix,
		configFilePath: DefaultConfigFilePath,
		viper:          viper.New(),
	}
}

// SetConfigFilePath sets the path to the configuration file
func (l *Load) SetConfigFilePath(path string) error {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext != "yaml" && ext != "yml" && ext != "json" {
		return fmt.Errorf("unsupported config file extension: %s", ext)
	}
	l.configFilePath = path
	l.configFileExt = ext
	return nil
}

// Load reads the configuration from the file and environment variables
func (l *Load) Load() (*Config, error) {
	l.setViperDefaults()
	l.prepareViper()

	if err := l.loadFromFile(); err != nil {
		return l.cfg, err
	}
	if err := l.viperToCfg(); err != nil {
		return l.cfg, err
	}

	slog.Info("Config loaded successfully")
	return l.cfg, nil
}

func (l *Load) setViperDefaults() {
	defaultsMap := map[string]interface{}{}
	if err := mapstructure.Decode(NewConfig(), &defaultsMap); err != nil {
		slog.Errorf("error while setting defaults: %v", err)
		return
	}
	for k, v := range defaultsMap {
		l.viper.SetDefault(k, v)
	}
}

func (l *Load) prepareViper() {
	l.viper.SetEnvPrefix(l.envPrefix)
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.viper.AutomaticEnv()
}

func (l *Load) loadFromFile() error {
	if _, err := os.Stat(l.configFilePath); os.IsNotExist(err) {
		slog.Warnf("Config file not found at %s — using defaults", l.configFilePath)
		return nil
	}

	l.viper.SetConfigFile(l.configFilePath)
	if err := l.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	slog.Infof("Loaded config from file: %s", l.configFilePath)
	return nil
}

func (l *Load) viperToCfg() error {
	if err := l.viper.Unmarshal(l.cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}
	l.updateViperFromConfig()
	return nil
}

func (l *Load) updateViperFromConfig() {
	configMap := map[string]any{}
	if err := mapstructure.Decode(l.cfg, &configMap); err != nil {
		slog.Errorf("failed to encode config back to map: %v", err)
		return
	}
	for key, value := range configMap {
		l.viper.Set(key, value)
	}
}
