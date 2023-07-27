package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CacheLimit int    `yaml:"cacheLimit"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	CacheDir   string `yaml:"cacheDir"`
}

func LoadConfig(configPath string) (*Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}

	conf.checkDefaults()

	err = os.MkdirAll(conf.CacheDir, 0o755)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *Config) Address() string {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 80
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) checkDefaults() {
	if c.CacheLimit < 1 {
		c.CacheLimit = 10
	}

	if c.CacheDir == "" {
		c.CacheDir = "tmp"
	}

	if c.Host == "" {
		c.Host = "0.0.0.0"
	}

	if c.Port < 1 {
		c.Port = 80
	}
}
