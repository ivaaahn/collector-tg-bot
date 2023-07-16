package config

import "collector/internal/infra"

type BotConfig struct {
	Token string
}

type AppConfig struct {
	DatabaseConfig *infra.DatabaseConfig `toml:"database"`
	BotConfig      *BotConfig            `toml:"bot"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}
