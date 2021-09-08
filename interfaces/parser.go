package interfaces

import (
	"errors"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"
	
	"github.com/prometheus/common/log"
)
// Parse parses cli output and tries to find interfaces with related stats
func (c *interfaceCollector) Parse(ostype string, output string) ([]Interface, error) {
	log.Debugf("OS: %s\n", ostype)
	switch ostype {
	case rpc.ArubaSwitch:
		return c.ParseArubaSwitch(ostype, output)
	case rpc.ArubaCXSwitch:
		return c.ParseArubaSwitch(ostype, output)
	default:
		return nil, errors.New("'show interface' is not implemented for " + ostype)
	}
}

// Parse parses ArubaSwitch cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaSwitch(ostype string, output string) ([]Interface, error) {
	items := []Interface{}
	newIfRegexp := regexp.MustCompile(`^\s+Status and Counters - Port Counters for (?:trunk|port) ((?:Trk)?\d+\/?\d*)\s*$`)
	descRegexp := regexp.MustCompile(`^\s+Name\s+:\s+(.*?)\s*$`)
	macRegexp := regexp.MustCompile(`^\s+MAC Address\s+:\s+(.*?)\s*$`)
	linkStatusRegexp := regexp.MustCompile(`^\s+Link Status\s+:\s+(Up|Down)\s*$`)
	portEnabledRegexp := regexp.MustCompile(`^\s+Port Enabled\s+:\s+(Yes|No)\s*$`)
	bytesRegexp := regexp.MustCompile(`\s+Bytes Rx\s+:\s+(.*?)\s+Bytes Tx\s+:\s+(.*?)\s*$`)
	unicastRegexp := regexp.MustCompile(`\s+Unicast Rx\s+:\s+(.*?)\s+Unicast Tx\s+:\s+(.*?)\s*$`)
	BandMcastRegexp := regexp.MustCompile(`\s+Bcast\/Mcast Rx\s+:\s+(.*?)\s+Bcast\/Mcast Tx\s+:\s+(.*?)\s*$`)
	RxDropsRegexp := regexp.MustCompile(`\s+Discard Rx\s+:\s+(.*?)\s+Out Queue Len\s+:\s+(.*?)\s*$`)
	TxDropsRegexp := regexp.MustCompile(`\s+FCS Rx\s+:\s+(.*?)\s+Drops Tx\s+:\s+(.*?)\s*$`)
	RxErrorsRegexp := regexp.MustCompile(`\s+Total Rx Errors\s+:\s+(.*?)\s+Deferred Tx\s+:\s+(.*?)\s*$`)
	TxLateCollnRegexp := regexp.MustCompile(`\s+Runts Rx\s+:\s+(.*?)\s+Late Colln Tx\s+:\s+(.*?)\s*$`)
	TxExcessCollnRegexp := regexp.MustCompile(`\s+Giants Rx\s+:\s+(.*?)\s+Excessive Colln\s+:\s+(.*?)\s*$`)
	
	current := Interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if current != (Interface{}) {
				items = append(items, current)
			}
			current = Interface{
				Name: matches[1],
				Description: "",
				MacAddress: "",
				OperStatus: "down",
				AdminStatus: "down",
				RxBytes: 0,
				TxBytes: 0,
				RxPackets: 0,
				TxPackets: 0,
				RxUnicast: 0,
				TxUnicast: 0,
				RxBcast: 0,
				TxBcast: 0,
				RxMcast: 0,
				TxMcast: 0,
				RxBandMcast: 0,
				TxBandMcast: 0,
				RxDrops: 0,
				TxDrops: 0,
				RxErrors: 0,
				TxErrors: 0,
			}
			continue
		}

		if matches := descRegexp.FindStringSubmatch(line); matches != nil {
			current.Description = matches[1]
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			current.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := linkStatusRegexp.FindStringSubmatch(line); matches != nil {
			if strings.ToLower(matches[1]) == "up" {
				current.OperStatus = "up"
			}
			continue
		}

		if matches := portEnabledRegexp.FindStringSubmatch(line); matches != nil {
			if strings.ToLower(matches[1]) == "yes" {
				current.AdminStatus = "up"
			}
			continue
		}

		if matches := bytesRegexp.FindStringSubmatch(line); matches != nil {
			current.RxBytes += util.Str2float64(matches[1])
			current.TxBytes += util.Str2float64(matches[2])
			continue
		}

		if matches := unicastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxUnicast += util.Str2float64(matches[1])
			current.RxPackets += util.Str2float64(matches[1])
			current.TxUnicast += util.Str2float64(matches[2])
			current.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := BandMcastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxBandMcast += util.Str2float64(matches[1])
			current.RxPackets += util.Str2float64(matches[1])
			current.TxBandMcast += util.Str2float64(matches[2])
			current.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := RxDropsRegexp.FindStringSubmatch(line); matches != nil {
			current.RxDrops += util.Str2float64(matches[1])
			continue
		}

		if matches := TxDropsRegexp.FindStringSubmatch(line); matches != nil {
			current.TxDrops += util.Str2float64(matches[2])
			continue
		}

		if matches := RxErrorsRegexp.FindStringSubmatch(line); matches != nil {
			current.RxErrors += util.Str2float64(matches[1])
			continue
		}

		if matches := TxLateCollnRegexp.FindStringSubmatch(line); matches != nil {
			current.TxErrors += util.Str2float64(matches[2])
			continue
		}

		if matches := TxExcessCollnRegexp.FindStringSubmatch(line); matches != nil {
			current.TxErrors += util.Str2float64(matches[2])
			continue
		}

	}
	return append(items, current), nil
}

// Parse parses ArubaCXSwitch cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaCXSwitch(ostype string, output string) ([]Interface, error) {
	items := []Interface{}
	newIfRegexp := regexp.MustCompile(`^(?:Interface|Aggregate) ((?:vlan|lag)?\d+\/?\d*\/?\d*) is (up|down)`)
	descRegexp := regexp.MustCompile(`^\s+Description:\s+(.*?)\s*$`)
	macRegexp := regexp.MustCompile(`^\s+Hardware: Ethernet, MAC Address:\s+(.*?)\s*$`)
	adminStateRegexp := regexp.MustCompile(`^\s+Admin state is (up|down)\s*$`)
	packetsRegexp := regexp.MustCompile(`^\s+Packets\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	unicastRegexp := regexp.MustCompile(`^\s+Unicast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	McastRegexp := regexp.MustCompile(`^\s+Multicast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	BcastRegexp := regexp.MustCompile(`^\s+Broadcast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	bytesRegexp := regexp.MustCompile(`\s+Bytes\s+(\d+)\s+(\d+)\s+(\d+)`)
	dropsRegexp := regexp.MustCompile(`\s+Dropped\s+(\d+)\s+(\d+)\s+(\d+)`)
	errorsRegexp := regexp.MustCompile(`\s+Errors\s+(\d+)\s+(\d+)\s+(\d+)`)
	
	current := Interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if current != (Interface{}) {
				items = append(items, current)
			}
			log.Infoln(matches[1])
			current = Interface{
				Name: matches[1],
				Description: "",
				MacAddress: "",
				OperStatus: "down",
				AdminStatus: "down",
				RxBytes: 0,
				TxBytes: 0,
				RxPackets: 0,
				TxPackets: 0,
				RxUnicast: 0,
				TxUnicast: 0,
				RxBcast: 0,
				TxBcast: 0,
				RxMcast: 0,
				TxMcast: 0,
				RxBandMcast: 0,
				TxBandMcast: 0,
				RxDrops: 0,
				TxDrops: 0,
				RxErrors: 0,
				TxErrors: 0,
			}
			if strings.ToLower(matches[2]) == "up" {
				current.OperStatus = "up"
			}
			continue
		}

		if matches := adminStateRegexp.FindStringSubmatch(line); matches != nil {
			if strings.ToLower(matches[1]) == "yes" {
				current.AdminStatus = "up"
			}
			continue
		}

		if matches := descRegexp.FindStringSubmatch(line); matches != nil {
			current.Description = matches[1]
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			current.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := packetsRegexp.FindStringSubmatch(line); matches != nil {
			current.RxPackets += util.Str2float64(matches[1])
			current.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := unicastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxUnicast += util.Str2float64(matches[1])
			current.TxUnicast += util.Str2float64(matches[2])
			continue
		}

		if matches := McastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxMcast += util.Str2float64(matches[1])
			current.TxMcast += util.Str2float64(matches[2])
			continue
		}

		if matches := BcastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxBcast += util.Str2float64(matches[1])
			current.TxBcast += util.Str2float64(matches[2])
			continue
		}

		if matches := bytesRegexp.FindStringSubmatch(line); matches != nil {
			current.RxBytes += util.Str2float64(matches[1])
			current.TxBytes += util.Str2float64(matches[2])
			continue
		}

		if matches := dropsRegexp.FindStringSubmatch(line); matches != nil {
			current.RxDrops += util.Str2float64(matches[1])
			current.TxDrops += util.Str2float64(matches[2])
			continue
		}

		if matches := errorsRegexp.FindStringSubmatch(line); matches != nil {
			current.RxErrors += util.Str2float64(matches[1])
			current.TxErrors += util.Str2float64(matches[1])
			continue
		}

	}
	return append(items, current), nil
}