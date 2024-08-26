package exporter

import (
	"errors"

	"github.com/pglomba/udpinger/pkg/client"
	"github.com/pglomba/udpinger/pkg/config"
)

type Exporter interface {
	Run(resultsCh <-chan client.ConvertedRTTCheckResult)
}

func NewExporter(config *config.Config) (Exporter, error) {
	switch config.ExporterType {
	case "stdout":
		return &StdoutExporter{}, nil
	case "prometheus":
		return &PrometheusExporter{
			address: config.PrometheusAddress,
		}, nil
	default:
		return nil, errors.New("invalid exporter type")
	}
}
