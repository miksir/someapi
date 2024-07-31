package config

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type DBConfig struct {
	ConnString string
}

type Config struct {
	DBMaster DBConfig
}

func IsDebug() bool {
	debug := flag.Bool("debug", false, "debug mode logging")
	flag.Parse()
	if debug != nil && *debug {
		return true
	}
	return false
}

func LoadConfig(logger zerolog.Logger) (*Config, error) {
	logger.Debug().Msg("loading config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("api")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	logger.Debug().Interface("cfg", cfg).Msg("loaded config")
	return &cfg, nil
}
