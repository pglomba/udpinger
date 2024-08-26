package exporter

import (
	"fmt"
	"log/slog"

	"github.com/pglomba/udpinger/pkg/client"
)

type StdoutExporter struct{}

func (e *StdoutExporter) Run(resultsCh <-chan client.ConvertedRTTCheckResult) {
	for {
		slog.Info("Starting STDOUT exporter")
		rttResult := <-resultsCh
		fmt.Printf("Target: %v, Results: %v, Min: %v, Max: %v, Avg: %v, Sent: %v, Received: %v, Packet Loss: %v\n",
			rttResult.Target,
			rttResult.Results,
			rttResult.Min,
			rttResult.Max,
			rttResult.Avg,
			rttResult.Sent,
			rttResult.Received,
			rttResult.PacketLoss,
		)
	}
}
