package system

import (
	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix string = "aruba_system_"

var (
	versionDesc     *prometheus.Desc
	uptimeDesc      *prometheus.Desc
	memoryTotalDesc *prometheus.Desc
	memoryUsedDesc  *prometheus.Desc
	memoryFreeDesc  *prometheus.Desc
	cpuUsedDesc     *prometheus.Desc
	cpuIdleDesc     *prometheus.Desc
)

func init() {
	l := []string{"target"}
	versionDesc = prometheus.NewDesc(prefix+"version", "Running OS version", append(l, "version"), nil)
	uptimeDesc = prometheus.NewDesc(prefix+"uptime", "Device uptime in seconds", append(l, "type"), nil)

	memoryTotalDesc = prometheus.NewDesc(prefix+"memory_total", "Total memory", append(l, "type"), nil)
	memoryUsedDesc = prometheus.NewDesc(prefix+"memory_used", "Used memory", append(l, "type"), nil)
	memoryFreeDesc = prometheus.NewDesc(prefix+"memory_free", "Free memory", append(l, "type"), nil)

	cpuUsedDesc = prometheus.NewDesc(prefix+"cpu_used_percent", "Percent CPU Used", append(l, "type"), nil)
	cpuIdleDesc = prometheus.NewDesc(prefix+"cpu_idle_percent", "Percent CPU Idle", append(l, "type"), nil)
}

type systemCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &systemCollector{}
}

// Name returns the name of the collector
func (*systemCollector) Name() string {
	return "System"
}

// Describe describes the metrics
func (*systemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- versionDesc
	ch <- uptimeDesc

	ch <- memoryTotalDesc
	ch <- memoryUsedDesc
	ch <- memoryFreeDesc

	ch <- cpuUsedDesc
	ch <- cpuIdleDesc
}

// CollectVersion collects version informations from Aruba Devices
func (c *systemCollector) CollectVersion(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	out, err := client.RunCommand([]string{"show version"})
	if err != nil {
		return err
	}
	item, err := c.ParseVersion(client.OSType, out)
	if err != nil {
		return err
	}
	l := append(labelValues, item.Version)
	ch <- prometheus.MustNewConstMetric(versionDesc, prometheus.GaugeValue, 1, l...)
	return nil
}

// CollectUptime collects uptime informations from Aruba Devices
func (c *systemCollector) CollectUptime(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out string
		err error
	)
	switch client.OSType {
	case "ArubaSwitch":
		out, err = client.RunCommand([]string{"show uptime"})
		if err != nil {
			return err
		}
	case "ArubaCXSwitch":
		out, err = client.RunCommand([]string{"show uptime"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show version"})
		if err != nil {
			return err
		}
	}
	item, err := c.ParseUptime(client.OSType, out)
	if err != nil {
		return err
	}
	l := append(labelValues, item.Type)
	ch <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.GaugeValue, item.Uptime, l...)
	return nil
}

// CollectMemory collects memory informations from Aruba Devices
func (c *systemCollector) CollectMemory(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out string
		err error
	)
	switch client.OSType {
	case "ArubaSwitch":
		out, err = client.RunCommand([]string{"display memory"})
		if err != nil {
			return err
		}
	case "ArubaCXSwitch":
		out, err = client.RunCommand([]string{"top memory"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show memory"})
		if err != nil {
			return err
		}
	}
	items, err := c.ParseMemory(client.OSType, out)
	if err != nil {
		return err
	}
	for _, item := range items {
		l := append(labelValues, item.Type)
		ch <- prometheus.MustNewConstMetric(memoryTotalDesc, prometheus.GaugeValue, item.Total, l...)
		ch <- prometheus.MustNewConstMetric(memoryUsedDesc, prometheus.GaugeValue, item.Used, l...)
		ch <- prometheus.MustNewConstMetric(memoryFreeDesc, prometheus.GaugeValue, item.Free, l...)
	}
	return nil
}

// CollectCPU collects cpu informations from Aruba Devices
func (c *systemCollector) CollectCPU(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out string
		err error
	)
	switch client.OSType {
	case "ArubaController":
		out, err = client.RunCommand([]string{"show cpuload per-cpu"})
		if err != nil {
			return err
		}
	case "ArubaCXSwitch":
		out, err = client.RunCommand([]string{"show system"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show cpu"})
		if err != nil {
			return err
		}
	}
	items, err := c.ParseCPU(client.OSType, out)
	if err != nil {
		return err
	}
	for _, item := range items {
		l := append(labelValues, item.Type)
		ch <- prometheus.MustNewConstMetric(cpuUsedDesc, prometheus.GaugeValue, item.Used, l...)
		ch <- prometheus.MustNewConstMetric(cpuIdleDesc, prometheus.GaugeValue, item.Idle, l...)
	}
	return nil
}

// Collect collects metrics from Aruba Devices
func (c *systemCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	log.Debugf("client: %+v", client)
	log.Debugf("labelValues %+v", labelValues)

	err := c.CollectVersion(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectVersion for %s: %s\n", labelValues[0], err.Error())
	}
	err = c.CollectUptime(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectUptime for %s: %s\n", labelValues[0], err.Error())
	}
	err = c.CollectMemory(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectMemory for %s: %s\n", labelValues[0], err.Error())
	}
	err = c.CollectCPU(client, ch, labelValues)
	if err != nil {
		log.Debugf("CollectCPU for %s: %s\n", labelValues[0], err.Error())
	}
	return nil
}
