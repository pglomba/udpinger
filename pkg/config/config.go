package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/viper"
)

func init() {
	configFile := flag.String("config", "", "Config file")
	flag.Parse()

	if *configFile == "" {
		slog.Error("Config file is not specified")
		os.Exit(1)
	}

	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

type Config struct {
	Listen            string   `mapstructure:"listen"`
	Interval          int      `mapstructure:"interval"`
	Count             int      `mapstructure:"count"`
	Timeout           int      `mapstructure:"timeout"`
	Targets           []string `mapstructure:"targets"`
	Unit              string   `mapstructure:"unit"`
	ExporterType      string   `mapstructure:"exporter_type"`
	PrometheusAddress string   `mapstructure:"prometheus_address"`
	Debug             bool     `mapstructure:"debug"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	fmt.Println(cfg)
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Interval == 0 {
		c.Interval = 1
	}

	if c.Count == 0 {
		c.Count = 3
	}

	if c.Timeout == 0 {
		c.Timeout = 10
	}

	if len(c.Targets) == 0 {
		return errors.New("no targets specified")
	}

	validUnits := []string{"ns", "us", "ms", ""}
	if !slices.Contains(validUnits, c.Unit) {
		return errors.New("invalid unit specified")
	} else if c.Unit == "" {
		c.Unit = "ms"
	}

	if c.ExporterType == "" {
		c.ExporterType = "stdout"
	}

	if c.PrometheusAddress == "" {
		c.PrometheusAddress = "127.0.0.1:2112"
	}

	return nil
}
