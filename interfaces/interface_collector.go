package interfaces

import (
	"github.com/yankiwi/aruba_exporter/collector"
	"github.com/yankiwi/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix string = "aruba_interface_"

var (
	rxBytesDesc     *prometheus.Desc
	rxPacketsDesc   *prometheus.Desc
	rxErrorsDesc    *prometheus.Desc
	rxDropsDesc     *prometheus.Desc
	rxUnicastDesc   *prometheus.Desc
	rxBcastDesc     *prometheus.Desc
	rxMcastDesc     *prometheus.Desc
	rxBandMcastDesc *prometheus.Desc

	txBytesDesc     *prometheus.Desc
	txPacketsDesc   *prometheus.Desc
	txErrorsDesc    *prometheus.Desc
	txDropsDesc     *prometheus.Desc
	txUnicastDesc   *prometheus.Desc
	txBcastDesc     *prometheus.Desc
	txMcastDesc     *prometheus.Desc
	txBandMcastDesc *prometheus.Desc

	adminStatusDesc *prometheus.Desc
	operStatusDesc  *prometheus.Desc
	errorStatusDesc *prometheus.Desc
)

func init() {
	l := []string{"target", "name", "description", "mac"}

	rxBytesDesc = prometheus.NewDesc(prefix+"rx_bytes", "Received data in bytes", l, nil)
	rxPacketsDesc = prometheus.NewDesc(prefix+"rx_packets", "Number of incoming packets", l, nil)
	rxErrorsDesc = prometheus.NewDesc(prefix+"rx_errors", "Number of errors caused by incoming packets", l, nil)
	rxDropsDesc = prometheus.NewDesc(prefix+"rx_drops", "Number of dropped incoming packets", l, nil)
	rxUnicastDesc = prometheus.NewDesc(prefix+"rx_unicast", "Received unicast packets", l, nil)
	rxBcastDesc = prometheus.NewDesc(prefix+"rx_broadcast", "Received broadcast packets", l, nil)
	rxMcastDesc = prometheus.NewDesc(prefix+"rx_multicast", "Received multicast packets", l, nil)
	rxBandMcastDesc = prometheus.NewDesc(prefix+"rx_broadcast_and_multicast", "Received number of combined broadcast and multicast packets", l, nil)

	txBytesDesc = prometheus.NewDesc(prefix+"tx_bytes", "Transmitted data in bytes", l, nil)
	txPacketsDesc = prometheus.NewDesc(prefix+"tx_packets", "Number of outgoing packets", l, nil)
	txErrorsDesc = prometheus.NewDesc(prefix+"tx_errors", "Number of errors caused by outgoing packets", l, nil)
	txDropsDesc = prometheus.NewDesc(prefix+"tx_drops", "Number of dropped outgoing packets", l, nil)
	txUnicastDesc = prometheus.NewDesc(prefix+"tx_unicast", "Transmitted unicast packets", l, nil)
	txBcastDesc = prometheus.NewDesc(prefix+"tx_broadcast", "Transmitted broadcast packets", l, nil)
	txMcastDesc = prometheus.NewDesc(prefix+"tx_multicast", "Transmitted multicast packets", l, nil)
	txBandMcastDesc = prometheus.NewDesc(prefix+"tx_broadcast_and_multicast", "Transmitted number of combined broadcast and multicast packets", l, nil)

	adminStatusDesc = prometheus.NewDesc(prefix+"admin_up", "Admin operational status", l, nil)
	operStatusDesc = prometheus.NewDesc(prefix+"up", "Interface operational status", l, nil)
	errorStatusDesc = prometheus.NewDesc(prefix+"error_status", "Admin and operational status differ", l, nil)
}

type interfaceCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &interfaceCollector{}
}

// Name returns the name of the collector
func (*interfaceCollector) Name() string {
	return "Interfaces"
}

// Describe describes the metrics
func (*interfaceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- rxBytesDesc
	ch <- rxPacketsDesc
	ch <- rxErrorsDesc
	ch <- rxDropsDesc
	ch <- rxUnicastDesc
	ch <- rxBcastDesc
	ch <- rxMcastDesc
	ch <- rxBandMcastDesc

	ch <- txBytesDesc
	ch <- txPacketsDesc
	ch <- txDropsDesc
	ch <- txErrorsDesc
	ch <- txUnicastDesc
	ch <- txBcastDesc
	ch <- txMcastDesc
	ch <- txBandMcastDesc

	ch <- adminStatusDesc
	ch <- operStatusDesc
	ch <- errorStatusDesc
}

// Collect collects metrics from Aruba
func (c *interfaceCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out string
		items []Interface
		err error
	)
	switch client.OSType {
	case "ArubaSwitch":
		out, err = client.RunCommand([]string{"show interfaces ethernet all"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show interfaces"})
		if err != nil {
			return err
		}
		log.Warnf("Interfaces parsing not available for %s\n", client.OSType)
		return nil
	}
	items, err = c.Parse(client.OSType, out)
	if err != nil {
		log.Warnf("Parse interfaces failed for %s: %s\n", labelValues[0], err.Error())
		return nil
	}

	for _, item := range items {
		l := append(labelValues, item.Name, item.Description, item.MacAddress)

		errorStatus := 0
		if item.AdminStatus != item.OperStatus {
			errorStatus = 1
		}
		adminStatus := 0
		if item.AdminStatus == "up" {
			adminStatus = 1
		}
		operStatus := 0
		if item.OperStatus == "up" {
			operStatus = 1
		}
		ch <- prometheus.MustNewConstMetric(rxBytesDesc, prometheus.CounterValue, item.RxBytes, l...)
		ch <- prometheus.MustNewConstMetric(rxPacketsDesc, prometheus.CounterValue, item.RxPackets, l...)
		ch <- prometheus.MustNewConstMetric(rxErrorsDesc, prometheus.CounterValue, item.RxErrors, l...)
		ch <- prometheus.MustNewConstMetric(rxDropsDesc, prometheus.CounterValue, item.RxDrops, l...)
		ch <- prometheus.MustNewConstMetric(rxUnicastDesc, prometheus.CounterValue, item.RxUnicast, l...)
		ch <- prometheus.MustNewConstMetric(rxBcastDesc, prometheus.CounterValue, item.RxBcast, l...)
		ch <- prometheus.MustNewConstMetric(rxMcastDesc, prometheus.CounterValue, item.RxMcast, l...)
		ch <- prometheus.MustNewConstMetric(rxBandMcastDesc, prometheus.CounterValue, item.RxBandMcast, l...)

		ch <- prometheus.MustNewConstMetric(txBytesDesc, prometheus.CounterValue, item.TxBytes, l...)
		ch <- prometheus.MustNewConstMetric(txPacketsDesc, prometheus.CounterValue, item.TxPackets, l...)
		ch <- prometheus.MustNewConstMetric(txErrorsDesc, prometheus.CounterValue, item.TxErrors, l...)
		ch <- prometheus.MustNewConstMetric(txDropsDesc, prometheus.CounterValue, item.TxDrops, l...)
		ch <- prometheus.MustNewConstMetric(txUnicastDesc, prometheus.CounterValue, item.TxUnicast, l...)
		ch <- prometheus.MustNewConstMetric(txBcastDesc, prometheus.CounterValue, item.TxBcast, l...)
		ch <- prometheus.MustNewConstMetric(txMcastDesc, prometheus.CounterValue, item.TxMcast, l...)
		ch <- prometheus.MustNewConstMetric(txBandMcastDesc, prometheus.CounterValue, item.TxBandMcast, l...)

		ch <- prometheus.MustNewConstMetric(adminStatusDesc, prometheus.GaugeValue, float64(adminStatus), l...)
		ch <- prometheus.MustNewConstMetric(operStatusDesc, prometheus.GaugeValue, float64(operStatus), l...)
		ch <- prometheus.MustNewConstMetric(errorStatusDesc, prometheus.GaugeValue, float64(errorStatus), l...)
	}

	return nil
}