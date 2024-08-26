package exporter

import (
	"log/slog"
	"net/http"

	"github.com/pglomba/udpinger/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusExporter struct {
	address string
}

type Metrics struct {
	packetLoss *prometheus.GaugeVec
	avgRtt     *prometheus.GaugeVec
	minRtt     *prometheus.GaugeVec
	maxRtt     *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		avgRtt: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "udpinger_average_rtt",
			Help: "Average round trip time.",
		}, []string{"target"}),
		minRtt: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "udpinger_min_rtt",
			Help: "Minimum round trip time.",
		}, []string{"target"}),
		maxRtt: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "udpinger_max_rtt",
			Help: "Maximum round trip time.",
		}, []string{"target"}),
		packetLoss: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "udpinger_packet_loss",
			Help: "Packet loss.",
		}, []string{"target"}),
	}
	reg.MustRegister(m.avgRtt, m.minRtt, m.maxRtt, m.packetLoss)
	return m
}

func (e *PrometheusExporter) Run(resultsCh <-chan client.ConvertedRTTCheckResult) {
	registry := prometheus.NewRegistry()
	metrics := NewMetrics(registry)
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	go func(handler http.Handler) {
		slog.Info("Starting Prometheus exporter on http://" + e.address + "/metrics")
		http.Handle("/metrics", promHandler)
		http.ListenAndServe(e.address, nil)

	}(promHandler)

	for {
		rttResult := <-resultsCh
		metrics.avgRtt.WithLabelValues(rttResult.Target).Set(float64(rttResult.Avg))
		metrics.minRtt.WithLabelValues(rttResult.Target).Set(float64(rttResult.Min))
		metrics.maxRtt.WithLabelValues(rttResult.Target).Set(float64(rttResult.Max))
		metrics.packetLoss.WithLabelValues(rttResult.Target).Set(rttResult.PacketLoss)
	}
}
