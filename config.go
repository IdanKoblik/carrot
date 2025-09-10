package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	InfluxdbConfig `yaml:"Influx"`
	Rabbit RabbitConfig `yaml:"Rabbit"`
	Api ApiConfig `yaml:"Api"`
}

type InfluxdbConfig struct {
	Url string `yaml:"url"`
	Token string `yaml:"token"`
	Org string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

type RabbitConfig struct {
	Channel string `yaml:"Channel"`
	Host string `yaml:"Host"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
	Port int `yaml:"Port"`
}

type ApiConfig struct {
	Host string `yaml:"Host"`
	Port int `yaml:"Port"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
