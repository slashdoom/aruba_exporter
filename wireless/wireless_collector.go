package wireless

import (
	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix string = "aruba_system_"

type wirelessCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &wirelessCollector{}
}

// Name returns the name of the collector
func (*wirelessCollector) Name() string {
	return "Wireless"
}

// Describe describes the metrics
func (*wirelessCollector) Describe(ch chan<- *prometheus.Desc) {

}

// Collect collects metrics from Aruba Devices
func (c *wirelessCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	log.Debugf("client: %+v", client)
	log.Debugf("labelValues %+v", labelValues)

	return nil
}