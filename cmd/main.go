package main

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"

	"github.com/pglomba/udpinger/pkg/client"
	"github.com/pglomba/udpinger/pkg/exporter"
	"github.com/pglomba/udpinger/pkg/server"
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

type config struct {
	Port     int      `mapstructure:"port"`
	Interval int      `mapstructure:"interval"`
	Count    int      `mapstructure:"count"`
	Timeout  int      `mapstructure:"timeout"`
	Targets  []string `mapstructure:"targets"`
	Unit     string   `mapstructure:"unit"`
}

func newConfig() (*config, error) {
	var cfg config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *config) Validate() error {
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

	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := newConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	srv, err := server.NewServer(":"+strconv.Itoa(config.Port), 1024)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	errCh := make(chan error)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	resultsCh := make(chan client.ConvertedRTTCheckResult)
	go func() {
		output, err := exporter.NewExporter(exporter.File)
		if err != nil {
			slog.Error(err.Error())
		}
		output.Run(resultsCh)
	}()

	for _, target := range config.Targets {
		clt, err := client.NewClient(target)
		if err != nil {
			slog.Error(err.Error())
			continue
		} else {
			slog.Info("created client for target " + target)
		}

		go func() {
			clt.StartMonitor(config.Interval, config.Count, config.Timeout, resultsCh, config.Unit)
		}()
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case err, ok := <-errCh:
		if ok {
			slog.Error(err.Error())
		}
	case sig := <-signalCh:
		slog.Info("Signal " + sig.String() + " received")
	}
}
