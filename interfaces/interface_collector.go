package interfaces

import (
	"github.com/yankiwi/aruba_exporter/collector"
	"github.com/yankiwi/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

	txBytesDesc     *prometheus.Desc
	txPacketsDesc   *prometheus.Desc
	txErrorsDesc    *prometheus.Desc
	txDropsDesc     *prometheus.Desc
	txUnicastDesc   *prometheus.Desc
	txBcastDesc     *prometheus.Desc
	txMcastDesc     *prometheus.Desc

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

	txBytesDesc = prometheus.NewDesc(prefix+"tx_bytes", "Transmitted data in bytes", l, nil)
	txPacketsDesc = prometheus.NewDesc(prefix+"tx_packets", "Number of outgoing packets", l, nil)
	txErrorsDesc = prometheus.NewDesc(prefix+"tx_errors", "Number of errors caused by outgoing packets", l, nil)
	txDropsDesc = prometheus.NewDesc(prefix+"tx_drops", "Number of dropped outgoing packets", l, nil)
	txUnicastDesc = prometheus.NewDesc(prefix+"tx_unicast", "Transmitted unicast packets", l, nil)
	txBcastDesc = prometheus.NewDesc(prefix+"tx_broadcast", "Transmitted broadcast packets", l, nil)
	txMcastDesc = prometheus.NewDesc(prefix+"tx_multicast", "Transmitted multicast packets", l, nil)

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

	ch <- txBytesDesc
	ch <- txPacketsDesc
	ch <- txDropsDesc
	ch <- txErrorsDesc
	ch <- txUnicastDesc
	ch <- txBcastDesc
	ch <- txMcastDesc

	ch <- adminStatusDesc
	ch <- operStatusDesc
	ch <- errorStatusDesc
}

// Collect collects metrics from Aruba
func (c *interfaceCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		out   string
		items map[string]Interface
		err   error
	)

	switch client.OSType {
	case "ArubaInstant":
		out, err = client.RunCommand([]string{"show interface counters"})
		if err != nil {
			return err
		}
	case "ArubaSwitch":
		out, err = client.RunCommand([]string{"show interfaces ethernet all","display interface"})
		if err != nil {
			return err
		}
	default:
		out, err = client.RunCommand([]string{"show interface"})
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

	for intName, intData := range items {
		l := append(labelValues, intName, intData.Description, intData.MacAddress)

		errorStatus := 0
		if intData.AdminStatus != intData.OperStatus {
			errorStatus = 1
		}
		adminStatus := 0
		if intData.AdminStatus == "up" {
			adminStatus = 1
		}
		operStatus := 0
		if intData.OperStatus == "up" {
			operStatus = 1
		}
		ch <- prometheus.MustNewConstMetric(rxBytesDesc, prometheus.CounterValue, intData.RxBytes, l...)
		ch <- prometheus.MustNewConstMetric(rxPacketsDesc, prometheus.CounterValue, intData.RxPackets, l...)
		ch <- prometheus.MustNewConstMetric(rxErrorsDesc, prometheus.CounterValue, intData.RxErrors, l...)
		ch <- prometheus.MustNewConstMetric(rxDropsDesc, prometheus.CounterValue, intData.RxDrops, l...)
		ch <- prometheus.MustNewConstMetric(rxUnicastDesc, prometheus.CounterValue, intData.RxUnicast, l...)
		ch <- prometheus.MustNewConstMetric(rxBcastDesc, prometheus.CounterValue, intData.RxBcast, l...)
		ch <- prometheus.MustNewConstMetric(rxMcastDesc, prometheus.CounterValue, intData.RxMcast, l...)

		ch <- prometheus.MustNewConstMetric(txBytesDesc, prometheus.CounterValue, intData.TxBytes, l...)
		ch <- prometheus.MustNewConstMetric(txPacketsDesc, prometheus.CounterValue, intData.TxPackets, l...)
		ch <- prometheus.MustNewConstMetric(txErrorsDesc, prometheus.CounterValue, intData.TxErrors, l...)
		ch <- prometheus.MustNewConstMetric(txDropsDesc, prometheus.CounterValue, intData.TxDrops, l...)
		ch <- prometheus.MustNewConstMetric(txUnicastDesc, prometheus.CounterValue, intData.TxUnicast, l...)
		ch <- prometheus.MustNewConstMetric(txBcastDesc, prometheus.CounterValue, intData.TxBcast, l...)
		ch <- prometheus.MustNewConstMetric(txMcastDesc, prometheus.CounterValue, intData.TxMcast, l...)

		ch <- prometheus.MustNewConstMetric(adminStatusDesc, prometheus.GaugeValue, float64(adminStatus), l...)
		ch <- prometheus.MustNewConstMetric(operStatusDesc, prometheus.GaugeValue, float64(operStatus), l...)
		ch <- prometheus.MustNewConstMetric(errorStatusDesc, prometheus.GaugeValue, float64(errorStatus), l...)
	}

	return nil
}
