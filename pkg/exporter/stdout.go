package exporter

import (
	"fmt"

	"github.com/pglomba/udpinger/pkg/client"
)

type FileExporter struct{}

func (f *FileExporter) Run(resultsCh <-chan client.ConvertedRTTCheckResult) {
	for {
		rttCheckResult := <-resultsCh

		fmt.Printf("Target: %v, Results: %v, Min: %v, Max: %v, Avg: %v, Sent: %v, Received: %v, Packet Loss: %v, Timestamp: %v\n",
			rttCheckResult.Target,
			rttCheckResult.Results,
			rttCheckResult.Min,
			rttCheckResult.Max,
			rttCheckResult.Avg,
			rttCheckResult.Sent,
			rttCheckResult.Received,
			rttCheckResult.PacketLoss,
			rttCheckResult.Timestamp,
		)
	}
}
