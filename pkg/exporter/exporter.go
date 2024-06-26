package exporter

import (
	"errors"

	"github.com/pglomba/udpinger/pkg/client"
)

type ExporterType int

const (
	File ExporterType = iota
)

type Exporter interface {
	Run(resultsCh <-chan client.ConvertedRTTCheckResult)
}

func NewExporter(ExporterType ExporterType) (Exporter, error) {
	switch ExporterType {
	case File:
		return &FileExporter{}, nil
	default:
		return nil, errors.New("invalid exporter type")
	}
}
