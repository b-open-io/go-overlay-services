package config

func (l *Load) applyDefaults() {
	defaultCfg := DefaultConfig()

	if l.cfg.Name == "" {
		l.cfg.Name = defaultCfg.Name
	}
	if l.cfg.PrivateKey == "" {
		l.cfg.PrivateKey = defaultCfg.PrivateKey
	}
	if l.cfg.HostingURL == "" {
		l.cfg.HostingURL = defaultCfg.HostingURL
	}
	if l.cfg.EngineConfig.ChainTracker == "" {
		l.cfg.EngineConfig.ChainTracker = defaultCfg.EngineConfig.ChainTracker
	}
	if l.cfg.EngineConfig.LogPrefix == "" {
		l.cfg.EngineConfig.LogPrefix = defaultCfg.EngineConfig.LogPrefix
	}
	if l.cfg.DatabaseConfig.URL == "" {
		l.cfg.DatabaseConfig.URL = defaultCfg.DatabaseConfig.URL
	}
	if l.cfg.MongoConfig.ConnectionString == "" {
		l.cfg.MongoConfig.ConnectionString = defaultCfg.MongoConfig.ConnectionString
	}
	if l.cfg.MigrationConfig.MigrationsDir == "" {
		l.cfg.MigrationConfig.MigrationsDir = defaultCfg.MigrationConfig.MigrationsDir
	}
	if l.cfg.LoggerConfig.Level == "" {
		l.cfg.LoggerConfig.Level = defaultCfg.LoggerConfig.Level
	}
	if l.cfg.LoggerConfig.Format == "" {
		l.cfg.LoggerConfig.Format = defaultCfg.LoggerConfig.Format
	}
}
