package config

import "fmt"

type redisConfig struct {
	Enable   bool   `mapstructure:"enable" yaml:"enable"`
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Password string `mapstructure:"password" yaml:"password"`
	Db       int    `mapstructure:"db" yaml:"db"`
}

func (c redisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
