package wireless

import (
	"fmt"

	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix string = "aruba_wireless_"

var (
	channelNoiseDesc *prometheus.Desc
	channelUtilDesc  *prometheus.Desc
	channelQualDesc  *prometheus.Desc
	channelCovrDesc  *prometheus.Desc
	channelIntfDesc  *prometheus.Desc
)

func init() {
	l := []string{"target", "channel", "band"}
	channelNoiseDesc = prometheus.NewDesc(prefix+"channel_noise", "Channel Noise", l, nil)
	channelUtilDesc = prometheus.NewDesc(prefix+"channel_utilization", "Channel Utilization", l, nil)
	channelQualDesc = prometheus.NewDesc(prefix+"channel_quailty", "Channel Quality", l, nil)
	channelCovrDesc = prometheus.NewDesc(prefix+"channel_coverage_index", "Channel Coverage Index", l, nil)
	channelIntfDesc = prometheus.NewDesc(prefix+"channel_interference_index", "Channel Interference Index", l, nil)
}

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

// CollectChannels collects memory informations from Aruba Devices
func (c *wirelessCollector) CollectChannels(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out string
		err error
	)
	switch client.OSType {
	case "ArubaController":
		out, err = client.RunCommand([]string{"show interface","show interface counters"})
		if err != nil {
			return err
		}
	case "ArubaInstant":
		out, err = client.RunCommand([]string{"show ap arm rf-summary"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show ap arm rf-summary"})
		if err != nil {
			return err
		}
	}
	items, err := c.ParseChannels(client.OSType, out)
	if err != nil {
		return err
	}
	for chName, chData := range items {
		l := append(labelValues, fmt.Sprintf("%v",chName), fmt.Sprintf("%v", chData.Band))

		ch <- prometheus.MustNewConstMetric(channelNoiseDesc, prometheus.GaugeValue, chData.Noise, l...)
		ch <- prometheus.MustNewConstMetric(channelUtilDesc, prometheus.GaugeValue, chData.ChUtil, l...)
		ch <- prometheus.MustNewConstMetric(channelQualDesc, prometheus.GaugeValue, chData.ChQual, l...)
		ch <- prometheus.MustNewConstMetric(channelCovrDesc, prometheus.GaugeValue, chData.CovrIndex, l...)
		ch <- prometheus.MustNewConstMetric(channelIntfDesc, prometheus.GaugeValue, chData.IntfIndex, l...)
	}
	return nil
}

// Collect collects metrics from Aruba Devices
func (c *wirelessCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	log.Debugf("client: %+v", client)
	log.Debugf("labelValues: %+v", labelValues)

	err := c.CollectChannels(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectChannels for %s: %s\n", labelValues[0], err.Error())
	}

	return nil
}