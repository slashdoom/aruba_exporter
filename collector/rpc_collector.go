package collector

import (
	"github.com/yankiwi/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
)

// RPCCollector collects metrics from Aruba devices using rpc.Client
type RPCCollector interface {
	// Name returns an human readable name for logging and debugging purposes
	Name() string

	// Describe describes the metrics
	Describe(ch chan<- *prometheus.Desc)

	// Collect collects metrics from Aruba devices
	Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error
}
