package wireless

import (
	"errors"
	"fmt"

	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix string = "aruba_wireless_"

var (
	apUp         *prometheus.Desc
	apController *prometheus.Desc
	apClients    *prometheus.Desc

	channelNoiseDesc *prometheus.Desc
	channelUtilDesc  *prometheus.Desc
	channelQualDesc  *prometheus.Desc
	channelCovrDesc  *prometheus.Desc
	channelIntfDesc  *prometheus.Desc
)

func init() {
	l := []string{"target", "name"}
	apUp = prometheus.NewDesc(prefix+"ap_up", "Scrape of AP was successful", l, nil)
	apController = prometheus.NewDesc(prefix+"ap_controller", "AP is Virtual Controller", l, nil)
	apClients = prometheus.NewDesc(prefix+"ap_clients", "AP Connected Clients ", l, nil)
	l = []string{"target", "ap", "channel", "band"}
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
	ch <- apUp
	ch <- apController
	ch <- apClients

	ch <- channelNoiseDesc
	ch <- channelUtilDesc
	ch <- channelQualDesc
	ch <- channelCovrDesc
	ch <- channelIntfDesc
}

func (c *wirelessCollector) CollectAccessPoints(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) (map[string]WirelessAccessPoint, error) {
	var (
		out string
		aps map[string]WirelessAccessPoint
		err error
	)

	switch client.OSType {
	case "ArubaController":
		out, err = client.RunCommand([]string{"show summary"})
		if err != nil {
			return make(map[string]WirelessAccessPoint), err
		}
		aps, err = c.ParseAccessPoints(client.OSType, out)
	case "ArubaInstant":
		out, err = client.RunCommand([]string{"show summary"})
		if err != nil {
			return make(map[string]WirelessAccessPoint), err
		}
		aps, err = c.ParseAccessPoints(client.OSType, out)
	default:
		err = errors.New("'CollectAccessPoints' is not implemented for " + client.OSType)
	}
	if err != nil {
		return make(map[string]WirelessAccessPoint), err
	}
	for apName, apData := range aps {
		l := append(labelValues, fmt.Sprintf("%v",apName))
		
		apUpStatus := 0
		if apData.Up {
			apUpStatus = 1
		}
		apVcStatus := 0
		if apData.Controller {
			apVcStatus = 1
		}

		ch <- prometheus.MustNewConstMetric(apUp, prometheus.GaugeValue, float64(apUpStatus), l...)
		ch <- prometheus.MustNewConstMetric(apController, prometheus.GaugeValue, float64(apVcStatus), l...)
		ch <- prometheus.MustNewConstMetric(apClients, prometheus.GaugeValue, apData.Clients, l...)
	}
	return aps, nil
}

// CollectChannels collects memory informations from Aruba Devices
func (c *wirelessCollector) CollectChannels(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) (map[string]WirelessRadio, error) {
	var (
		out string
		channels map[string]WirelessChannel
		radios map[string]WirelessRadio
		err error
	)

	switch client.OSType {
	case "ArubaController":
		out, err = client.RunCommand([]string{"show interface"})
		if err != nil {
			return make(map[string]WirelessRadio), err
		}
		channels, _, err = c.ParseChannels(client.OSType, out)
	case "ArubaInstant":
		out, err = client.RunCommand([]string{"show ap-env", "show ap arm rf-summary"})
		if err != nil {
			return make(map[string]WirelessRadio), err
		}
		channels, radios, err = c.ParseChannels(client.OSType, out)
	default:
		err = errors.New("'CollectChannels' is not implemented for " + client.OSType)
	}
	if err != nil {
		return make(map[string]WirelessRadio), err
	}
	for chChannel, chData := range channels {
		log.Debugf("channel data: %+v", chData)
		l := append(labelValues, fmt.Sprintf("%v", chData.AccessPoint), fmt.Sprintf("%v",chChannel), fmt.Sprintf("%v", chData.Band))

		ch <- prometheus.MustNewConstMetric(channelNoiseDesc, prometheus.GaugeValue, chData.NoiseFloor, l...)
		ch <- prometheus.MustNewConstMetric(channelUtilDesc, prometheus.GaugeValue, chData.ChUtil, l...)
		ch <- prometheus.MustNewConstMetric(channelQualDesc, prometheus.GaugeValue, chData.ChQual, l...)
		ch <- prometheus.MustNewConstMetric(channelCovrDesc, prometheus.GaugeValue, chData.CovrIndex, l...)
		ch <- prometheus.MustNewConstMetric(channelIntfDesc, prometheus.GaugeValue, chData.IntfIndex, l...)
	}
	return radios, nil
}

func (c *wirelessCollector) CollectRadios(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string, radios map[string]WirelessRadio) (error) {
	log.Debugf("client: %+v", client)
	log.Debugf("labelValues: %+v", labelValues)
	var (
		out string
		err error
	)

	switch client.OSType {
	case "ArubaController":
		out, err = client.RunCommand([]string{"show interface"})
		if err != nil {
			return err
		}
		radios, err = c.ParseRadios(client.OSType, radios, out)
	case "ArubaInstant":
		out, err = client.RunCommand([]string{"show ap monitor status"})
		if err != nil {
			return err
		}
		radios, err = c.ParseRadios(client.OSType, radios, out)
	default:
		err = errors.New("'CollectRadios' is not implemented for " + client.OSType)
	}
	if err != nil {
		return err
	}
	for radioId, radioData := range radios {
		log.Debugf("radio data: %+v", radioData)
		l := append(labelValues, fmt.Sprintf("%v", radioData.AccessPoint), fmt.Sprintf("%v",radioId), fmt.Sprintf("%v", radioData.Bssid))
		log.Debugf("channel labels: %+v", l)
		//ch <- prometheus.MustNewConstMetric(channelNoiseDesc, prometheus.GaugeValue, chData.NoiseFloor, l...)
		//ch <- prometheus.MustNewConstMetric(channelUtilDesc, prometheus.GaugeValue, chData.ChUtil, l...)
		//ch <- prometheus.MustNewConstMetric(channelQualDesc, prometheus.GaugeValue, chData.ChQual, l...)
		//ch <- prometheus.MustNewConstMetric(channelCovrDesc, prometheus.GaugeValue, chData.CovrIndex, l...)
		//ch <- prometheus.MustNewConstMetric(channelIntfDesc, prometheus.GaugeValue, chData.IntfIndex, l...)
	}
	return nil
}

// Collect collects metrics from Aruba Devices
func (c *wirelessCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	log.Debugf("client: %+v", client)
	log.Debugf("labelValues: %+v", labelValues)
	var err error
	
	var aps map[string]WirelessAccessPoint 
	aps, err = c.CollectAccessPoints(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectAccessPoints for %s: %s\n", labelValues[0], err.Error())
	}
	log.Debugf("aps: %+v", aps)

	var radios map[string]WirelessRadio 
	radios, err = c.CollectChannels(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectChannels for %s: %s\n", labelValues[0], err.Error())
	}
	log.Debugf("radios: %+v", radios)

	return nil
}