package config

import "github.com/spf13/viper"

type Config struct {
	WebSocketURL string `json:"websocketurl" yaml:"websocketurl"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
