package main

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Debug          bool            `mapstructure:"debug"`
	Storage        ConfigStorage   `mapstructure:"storage"`
	SyntaxDefault  string          `mapstructure:"syntax_default"`
	Syntax         []ConfigSyntax  `mapstructure:"syntax"`
	ExpiresDefault string          `mapstructure:"expires_default"`
	Expires        []ConfigExpires `mapstructure:"expires"`
}

type ConfigStorage struct {
	SQL ConfigStorageSQL `mapstructure:"sql"`
}

type ConfigStorageSQL struct {
	Host     string `mapstructure:"host"`
	DB       string `mapstructure:"db"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type ConfigSyntax struct {
	Label   string   `mapstructure:"label"`
	Syntax  string   `mapstructure:"syntax"`
	Aliases []string `mapstructure:"aliases"`
}

type ConfigExpires struct {
	Label   string `mapstructure:"label"`
	Expires int64  `mapstructure:"expires"`
}

func InitConfig() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetEnvPrefix("paste")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	setConfigDefaults()

	config := Config{}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func setConfigDefaults() {
	viper.SetDefault("debug", false)
	viper.SetDefault("storage.sql.host", "")
	viper.SetDefault("storage.sql.port", 3306)
	viper.SetDefault("storage.sql.db", "paste")
	viper.SetDefault("storage.sql.user", "")
	viper.SetDefault("storage.sql.password", "")
}
