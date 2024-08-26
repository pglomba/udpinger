package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pglomba/udpinger/pkg/client"
	"github.com/pglomba/udpinger/pkg/config"
	"github.com/pglomba/udpinger/pkg/exporter"
	"github.com/pglomba/udpinger/pkg/logger"
	"github.com/pglomba/udpinger/pkg/server"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.Debug)
	slog.SetDefault(log)

	srv, err := server.NewServer(cfg.Listen, 1024)
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
		metricsExporter, err := exporter.NewExporter(cfg)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		metricsExporter.Run(resultsCh)
	}()

	for _, target := range cfg.Targets {
		clt, err := client.NewClient(target)
		if err != nil {
			slog.Error(err.Error())
			continue
		} else {
			slog.Info("Created client for target " + target)
		}

		go func() {
			clt.StartMonitor(cfg.Interval, cfg.Count, cfg.Timeout, resultsCh, cfg.Unit)
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
