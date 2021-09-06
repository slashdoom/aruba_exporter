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
	if ostype != rpc.ArubaSwitch {
		return nil, errors.New("'show interface' is not implemented for " + ostype)
	}
	items := []Interface{}
	newIfRegexp := regexp.MustCompile(`^\s+Status and Counters - Port Counters for (?:trunk|port) ((?:Trk)?\d+\/?\d*)\s*$`)
	descRegexp := regexp.MustCompile(`^\s+Name\s+:\s+(.*?)\s*$`)
	macRegexp := regexp.MustCompile(`^\s+MAC Address\s+:\s+(.*?)\s*$`)
	linkStatusRegexp := regexp.MustCompile(`^\s+Link Status\s+:\s+(Up|Down)\s*$`)
	portEnabledRegexp := regexp.MustCompile(`^\s+Port Enabled\s+:\s+(Yes|No)\s*$`)
	bytesRegexp := regexp.MustCompile(`\s+Bytes Rx\s+:\s+(.*?)\s+Bytes Tx\s+:\s+(.*?)\s*$`)
	unicastRegexp := regexp.MustCompile(`\s+Unicast Rx\s+:\s+(.*?)\s+Unicast Tx\s+:\s+(.*?)\s*$`)
	BandMcastRegexp := regexp.MustCompile(`\s+Bcast\/Mcast Rx\s+:\s+(.*?)\s+Bcast\/Mcast Tx\s+:\s+(.*?)\s*$`)
	RxDrops := regexp.MustCompile(`\s+Discard Rx\s+:\s+(.*?)\s+Out Queue Len\s+:\s+(.*?)\s*$`)
	TxDrops := regexp.MustCompile(`\s+FCS Rx\s+:\s+(.*?)\s+Drops Tx\s+:\s+(.*?)\s*$`)
	RxErrors := regexp.MustCompile(`\s+Total Rx Errors\s+:\s+(.*?)\s+Deferred Tx\s+:\s+(.*?)\s*$`)
	TxLateColln := regexp.MustCompile(`\s+Runts Rx\s+:\s+(.*?)\s+Late Colln Tx\s+:\s+(.*?)\s*$`)
	TxExcessColln := regexp.MustCompile(`\s+Giants Rx\s+:\s+(.*?)\s+Excessive Colln\s+:\s+(.*?)\s*$`)
	
	current := Interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if current != (Interface{}) {
				items = append(items, current)
			}
			current = Interface{
				Name: matches[1],
				RxBytes: 0,
				TxBytes: 0,
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
			} else {
				current.OperStatus = "down"
			}
			continue
		}

		if matches := portEnabledRegexp.FindStringSubmatch(line); matches != nil {
			if strings.ToLower(matches[1]) == "up" {
				current.AdminStatus = "up"
			} else {
				current.AdminStatus = "down"
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
			current.TxUnicast += util.Str2float64(matches[2])
			continue
		}

		if matches := BandMcastRegexp.FindStringSubmatch(line); matches != nil {
			current.RxBandMcast += util.Str2float64(matches[1])
			current.TxBandMcast += util.Str2float64(matches[2])
			continue
		}

		if matches := RxDrops.FindStringSubmatch(line); matches != nil {
			current.RxDrops += util.Str2float64(matches[1])
			continue
		}

		if matches := TxDrops.FindStringSubmatch(line); matches != nil {
			current.TxDrops += util.Str2float64(matches[2])
			continue
		}

		if matches := RxErrors.FindStringSubmatch(line); matches != nil {
			current.RxErrors += util.Str2float64(matches[1])
			continue
		}

		if matches := TxLateColln.FindStringSubmatch(line); matches != nil {
			current.TxErrors += util.Str2float64(matches[2])
			continue
		}

		if matches := TxExcessColln.FindStringSubmatch(line); matches != nil {
			current.TxErrors += util.Str2float64(matches[2])
			continue
		}

	}
	return append(items, current), nil
}