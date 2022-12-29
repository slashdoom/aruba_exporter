package interfaces

import (
	"errors"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"

	log "github.com/sirupsen/logrus"
)

// Parse parses cli output and tries to find interfaces with related stats
func (c *interfaceCollector) Parse(ostype string, output string) (map[string]Interface, error) {
	log.Debugf("OS: %s\n", ostype)
	switch ostype {
	case rpc.ArubaController:
		return c.ParseArubaController(output)
	case rpc.ArubaInstant:
		return c.ParseArubaInstant(output)
	case rpc.ArubaSwitch: 
		return c.ParseArubaSwitch(output)
	case rpc.ArubaCXSwitch:
		return c.ParseArubaCXSwitch(output)
	default:
		return nil, errors.New("'show interface' is not implemented for " + ostype)
	}
}

// Parses ArubaController cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaController(output string) (map[string]Interface, error) {
	interfaces := make(map[string]Interface)

	newIfRegexp := regexp.MustCompile(`^GE (\d+\/\d+\/\d+) is (up|down), line protocol is (up|down)`)
	macRegexp := regexp.MustCompile(`^Hardware is.*, address is (.*?) \(bia (.*?)\)\s*$`)
	RxPacketsAndBytesRegexp := regexp.MustCompile(`^\s+(\d+)\spackets\sinput,\s(\d+) bytes\s*$`)
	RxErrors1Regexp := regexp.MustCompile(`^\s*Received\s\d+\sbroadcasts,\s(\d+)\srunts,\s(\d+)\sgiants,\s(\d+)\sthrottles\s*$`)
	RxErrors2Regexp := regexp.MustCompile(`^\s*\d+\sinput error bytes,\s(\d+)\sCRC,\s(\d+)\sframe\s*$`)
	TxPacketsAndBytesRegexp := regexp.MustCompile(`^\s*(\d+)\spackets output,\s(\d+)\sbytes\s*$`)
	TxErrors1Regexp := regexp.MustCompile(`^\s*\d+\soutput errors bytes,\s(\d+)\sdeferred\s*$`)
	TxErrors2Regexp := regexp.MustCompile(`^\s*(\d+)\scollisions,\s(\d+)\slate collisions,\s(\d+)\sthrottles\s*$`)

	p2InHeaderRegexp := regexp.MustCompile(`^\s*Port\s+InOctets`)
	p2OutHeaderRegexp := regexp.MustCompile(`^\s*Port\s+OutOctets`)
	p2IntRegexp := regexp.MustCompile(`^\s*^GE(\d+\/\d+\/\d+)\s+\d+\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)

	currentInt := Interface{}
	currentName := ""

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if currentInt != (Interface{}) {
				interfaces[currentName] = currentInt
			}
			log.Debugf("pass 1 interface: %+v", matches[1])
			currentName = matches[1]
			currentInt = Interface{
				Description: "",
				MacAddress:  "",
				OperStatus:  "down",
				AdminStatus: "down",
				RxBytes:     0,
				TxBytes:     0,
				RxPackets:   0,
				TxPackets:   0,
				RxUnicast:   0,
				TxUnicast:   0,
				RxBcast:     0,
				TxBcast:     0,
				RxMcast:     0,
				TxMcast:     0,
				RxDrops:     0,
				TxDrops:     0,
				RxErrors:    0,
				TxErrors:    0,
			}
			log.Debugf("AdminStatus: %+v", matches[2])
			if strings.ToLower(matches[2]) == "up" {
				currentInt.AdminStatus = "up"
			}
			log.Debugf("OperStatus: %+v", matches[3])
			if strings.ToLower(matches[3]) == "up" {
				currentInt.OperStatus = "up"
			}
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("MacAddress: %+v", matches[1])
			currentInt.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := RxPacketsAndBytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxPackets: %+v", matches[1])
			currentInt.RxPackets = util.Str2float64(matches[1])
			log.Debugf("RxBytes: %+v", matches[2])
			currentInt.RxBytes = util.Str2float64(matches[2])
			continue
		}

		if matches := RxErrors1Regexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxErrors1: %+v runts, %+v giants, %+v throttles", matches[1], matches[2], matches[3])
			currentInt.RxErrors += util.Str2float64(matches[1])
			currentInt.RxErrors += util.Str2float64(matches[2])
			currentInt.RxErrors += util.Str2float64(matches[3])
			continue
		}

		if matches := RxErrors2Regexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxErrors2: %+v CRC, %+v frame", matches[1], matches[2])
			currentInt.RxErrors += util.Str2float64(matches[1])
			currentInt.RxErrors += util.Str2float64(matches[2])
			continue
		}

		if matches := TxPacketsAndBytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxPackets: %+v", matches[1])
			currentInt.TxPackets = util.Str2float64(matches[1])
			log.Debugf("TxBytes: %+v", matches[2])
			currentInt.TxBytes = util.Str2float64(matches[2])
			continue
		}

		if matches := TxErrors1Regexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxErrors1: %+v deferred", matches[1])
			currentInt.TxErrors += util.Str2float64(matches[1])
			continue
		}

		if matches := TxErrors2Regexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxErrors2: %+v collisions, %+v late collisions, %+v throttles", matches[1], matches[2], matches[3])
			currentInt.TxErrors += util.Str2float64(matches[1])
			currentInt.TxErrors += util.Str2float64(matches[2])
			currentInt.TxErrors += util.Str2float64(matches[3])
			continue
		}
	}
	interfaces[currentName] = currentInt

	currentInt = Interface{}
	currentName = ""
	inputHeader := false
	outputHeader := false
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := p2InHeaderRegexp.FindStringSubmatch(line); matches != nil {
			inputHeader = true
			outputHeader = false
			log.Debugf("inputHeader: %+v", inputHeader)
			continue
		}
		if matches := p2OutHeaderRegexp.FindStringSubmatch(line); matches != nil {
			inputHeader = false
			outputHeader = true
			log.Debugf("outputHeader: %+v", outputHeader)
			continue
		}
		if inputHeader {
			if matches := p2IntRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("pass 2 in interface: %+v", matches[1])
				currentName = matches[1]
				currentInt = interfaces[currentName]
				log.Debugf("RxUnicast: %+v", matches[2])
				log.Debugf("RxMcast: %+v", matches[3])
				log.Debugf("RxBcast: %+v", matches[4])
				currentInt.RxUnicast = util.Str2float64(matches[2])
				currentInt.RxBcast = util.Str2float64(matches[3])
				currentInt.RxMcast = util.Str2float64(matches[4])
				interfaces[currentName] = currentInt
				continue
			}
		}
		if outputHeader {
			if matches := p2IntRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("pass 2 out interface: %+v", matches[1])
				currentName = matches[1]
				currentInt = interfaces[currentName]
				log.Debugf("TxUnicast: %+v", matches[2])
				log.Debugf("TxBcast: %+v", matches[3])
				log.Debugf("TxMcast: %+v", matches[4])
				currentInt.TxUnicast = util.Str2float64(matches[2])
				currentInt.TxBcast = util.Str2float64(matches[3])
				currentInt.TxMcast = util.Str2float64(matches[4])
				interfaces[currentName] = currentInt
				continue
			}
		}
	}
	
	return interfaces, nil
}

// Parse parses ArubaInstant cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaInstant(output string) (map[string]Interface, error) {
	interfaces := make(map[string]Interface)

	newIfRegexp := regexp.MustCompile(`^\s*(eth\d+) is (up|down), line protocol is (up|down)\s*$`)
	macRegexp := regexp.MustCompile(`^Hardware is.*, address is (.*?)\s*$`)
	RxPacketsRegexp := regexp.MustCompile(`^Received packets\s+(\d+)\s*$`)
	RxBytesRegexp := regexp.MustCompile(`^Received bytes\s+(\d+)\s*$`)
	RxDropsRegexp := regexp.MustCompile(`^Receive dropped\s+(\d+)\s*$`)
	RxErrorsRegexp := regexp.MustCompile(`Receive errors\s+(\d+)\s*$`)
	TxPacketsRegexp := regexp.MustCompile(`^Transmitted packets\s+(\d+)\s*$`)
	TxBytesRegexp := regexp.MustCompile(`^Transmitted bytes\s+(\d+)\s*$`)
	TxDropsRegexp := regexp.MustCompile(`^Transmitted dropped\s+(\d+)\s*$`)
	TxErrorsRegexp := regexp.MustCompile(`Transmitted errors\s+(\d+)\s*$`)

	currentInt := Interface{}
	currentName := ""

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if currentInt != (Interface{}) {
				interfaces[currentName] = currentInt
			}
			log.Debugf("interface: %+v", matches[1])
			currentName = matches[1]
			currentInt = Interface{
				Description: "",
				MacAddress:  "",
				OperStatus:  "down",
				AdminStatus: "down",
				RxBytes:     0,
				TxBytes:     0,
				RxPackets:   0,
				TxPackets:   0,
				RxUnicast:   0,
				TxUnicast:   0,
				RxBcast:     0,
				TxBcast:     0,
				RxMcast:     0,
				TxMcast:     0,
				RxDrops:     0,
				TxDrops:     0,
				RxErrors:    0,
				TxErrors:    0,
			}
			log.Debugf("AdminStatus: %+v", matches[2])
			if strings.ToLower(matches[2]) == "up" {
				currentInt.AdminStatus = "up"
			}
			log.Debugf("OperStatus: %+v", matches[3])
			if strings.ToLower(matches[3]) == "up" {
				currentInt.OperStatus = "up"
			}
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("MacAddress: %+v", matches[1])
			currentInt.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := RxPacketsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxPackets: %+v", matches[1])
			currentInt.RxPackets = util.Str2float64(matches[1])
			continue
		}

		if matches := RxBytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxBytes: %+v", matches[1])
			currentInt.RxBytes = util.Str2float64(matches[1])
			continue
		}

		if matches := RxDropsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxDrops: %+v", matches[1])
			currentInt.RxDrops = util.Str2float64(matches[1])
			continue
		}

		if matches := RxErrorsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxErrors: %+v", matches[1])
			currentInt.RxErrors = util.Str2float64(matches[1])
			continue
		}

		if matches := TxPacketsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxPackets: %+v", matches[1])
			currentInt.TxPackets = util.Str2float64(matches[1])
			continue
		}

		if matches := TxBytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxBytes: %+v", matches[1])
			currentInt.TxBytes = util.Str2float64(matches[1])
			continue
		}

		if matches := TxDropsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxDrops: %+v", matches[1])
			currentInt.TxDrops = util.Str2float64(matches[1])
			continue
		}

		if matches := TxErrorsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxErrors: %+v", matches[1])
			currentInt.TxErrors = util.Str2float64(matches[1])
			continue
		}
	}
	interfaces[currentName] = currentInt

	return interfaces, nil
}

// Parse parses ArubaSwitch cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaSwitch(output string) (map[string]Interface, error) {
	interfaces := make(map[string]Interface)

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

	p2newIfRegexp := regexp.MustCompile(`\s*((?:Trk)?\d+\/?\d*)\s+current state:\s+(UP|DOWN)\s*$`)
	p2inputTotalRegexp := regexp.MustCompile(`^\s*Input \(total\):\s+\d+ packets, \d+ bytes\s*$`)
	p2outputTotalRegexp := regexp.MustCompile(`^\s*Output \(total\):\s+\d+ packets, \d+ bytes\s*$`)
	p2PacketsRegexp := regexp.MustCompile(`^\s*\d+ unicasts, (\d+) broadcasts, (\d+) multicasts, \d+ pauses`)

	currentInt := Interface{}
	currentName := ""

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if currentInt != (Interface{}) {
				interfaces[currentName] = currentInt
			}
			log.Debugf("pass 1 interface: %+v", matches[1])
			currentName = matches[1]
			currentInt = Interface{
				Description: "",
				MacAddress:  "",
				OperStatus:  "down",
				AdminStatus: "down",
				RxBytes:     0,
				TxBytes:     0,
				RxPackets:   0,
				TxPackets:   0,
				RxUnicast:   0,
				TxUnicast:   0,
				RxBcast:     0,
				TxBcast:     0,
				RxMcast:     0,
				TxMcast:     0,
				RxDrops:     0,
				TxDrops:     0,
				RxErrors:    0,
				TxErrors:    0,
			}
			continue
		}

		if matches := descRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("Description: %+v", matches[1])
			currentInt.Description = matches[1]
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("MacAddress: %+v", matches[1])
			currentInt.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := linkStatusRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("OperStatus: %+v", matches[1])
			if strings.ToLower(matches[1]) == "up" {
				currentInt.OperStatus = "up"
			}
			continue
		}

		if matches := portEnabledRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("AdminStatus: %+v", matches[1])
			if strings.ToLower(matches[1]) == "yes" {
				currentInt.AdminStatus = "up"
			}
			continue
		}

		if matches := bytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxBytes: %+v", matches[1])
			log.Debugf("TxBytes: %+v", matches[2])
			currentInt.RxBytes += util.Str2float64(matches[1])
			currentInt.TxBytes += util.Str2float64(matches[2])
			continue
		}

		if matches := unicastRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxUnicast: %+v", matches[1])
			log.Debugf("TxUnicast: %+v", matches[2])
			currentInt.RxUnicast += util.Str2float64(matches[1])
			currentInt.RxPackets += util.Str2float64(matches[1])
			currentInt.TxUnicast += util.Str2float64(matches[2])
			currentInt.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := BandMcastRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxBandMcast: %+v", matches[1])
			log.Debugf("TxBandMcast: %+v", matches[2])
			currentInt.RxPackets += util.Str2float64(matches[1])
			currentInt.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := RxDropsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxDrops: %+v", matches[1])
			currentInt.RxDrops += util.Str2float64(matches[1])
			continue
		}

		if matches := TxDropsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxDrops: %+v", matches[2])
			currentInt.TxDrops += util.Str2float64(matches[2])
			continue
		}

		if matches := RxErrorsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxErrors: %+v", matches[1])
			currentInt.RxErrors += util.Str2float64(matches[1])
			continue
		}

		if matches := TxLateCollnRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxLateColln: %+v", matches[2])
			currentInt.TxErrors += util.Str2float64(matches[2])
			continue
		}

		if matches := TxExcessCollnRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("TxExcessColln: %+v", matches[2])
			currentInt.TxErrors += util.Str2float64(matches[2])
			continue
		}

	}
	interfaces[currentName] = currentInt

	currentInt = Interface{}
	currentName = ""
	inputTotalLine := false
	outputTotalLine := false
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := p2newIfRegexp.FindStringSubmatch(line); matches != nil {
			if currentInt != (Interface{}) {
				interfaces[currentName] = currentInt
			}
			log.Debugf("pass 2 interface: %+v", matches[1])
			currentName = matches[1]
			currentInt = interfaces[currentName]
			continue
		}
		if matches := p2inputTotalRegexp.FindStringSubmatch(line); matches != nil {
			inputTotalLine = true
			log.Debugf("inputTotalLine: %+v", inputTotalLine)
			continue
		}
		if matches := p2outputTotalRegexp.FindStringSubmatch(line); matches != nil {
			outputTotalLine = true
			log.Debugf("outputTotalLine: %+v", outputTotalLine)
			continue
		}
		if inputTotalLine {
			if matches := p2PacketsRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("RxBcast: %+v", matches[1])
				log.Debugf("RxMcast: %+v", matches[2])
				currentInt.RxBcast = util.Str2float64(matches[1])
				currentInt.RxMcast = util.Str2float64(matches[2])
				continue
			}
			inputTotalLine = false
		}
		if outputTotalLine {
			if matches := p2PacketsRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("TxBcast: %+v", matches[1])
				log.Debugf("TxMcast: %+v", matches[2])
				currentInt.TxBcast = util.Str2float64(matches[1])
				currentInt.TxMcast = util.Str2float64(matches[2])
				continue
			}
			outputTotalLine = false
		}
	}
	interfaces[currentName] = currentInt

	return interfaces, nil
}

// Parse parses ArubaCXSwitch cli output and tries to find interfaces with related stats
func (c *interfaceCollector) ParseArubaCXSwitch(output string) (map[string]Interface, error) {
	interfaces := make(map[string]Interface)
	newIfRegexp := regexp.MustCompile(`^(?:Interface|Aggregate) ((?:vlan|lag)?\d+\/?\d*\/?\d*) is (up|down)`)
	descRegexp := regexp.MustCompile(`^\s+Description:\s+(.*?)\s*$`)
	macRegexp := regexp.MustCompile(`^(?:\s+Hardware: Ethernet,)?\s+MAC Address\s*:\s+(.*?)\s*$`)
	adminStateRegexp := regexp.MustCompile(`^\s+Admin state is (up|down)\s*$`)
	packetsRegexp := regexp.MustCompile(`^\s+Packets\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	l3packetsRegexp := regexp.MustCompile(`^\s+L3 Packets\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	unicastRegexp := regexp.MustCompile(`^\s+Unicast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	McastRegexp := regexp.MustCompile(`^\s+Multicast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	BcastRegexp := regexp.MustCompile(`^\s+Broadcast\s+(\d+)\s+(\d+)\s+(\d+)\s*$`)
	bytesRegexp := regexp.MustCompile(`\s+Bytes\s+(\d+)\s+(\d+)\s+(\d+)`)
	l3bytesRegexp := regexp.MustCompile(`\s+L3 Bytes\s+(\d+)\s+(\d+)\s+(\d+)`)
	dropsRegexp := regexp.MustCompile(`\s+Dropped\s+(\d+)\s+(\d+)\s+(\d+)`)
	errorsRegexp := regexp.MustCompile(`\s+Errors\s+(\d+)\s+(\d+)\s+(\d+)`)

	currentInt := Interface{}
	currentName := ""

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		log.Tracef("line: %+v", line)
		if matches := newIfRegexp.FindStringSubmatch(line); matches != nil {
			if currentInt != (Interface{}) {
				interfaces[currentName] = currentInt
			}
			log.Debugf("interface: %+v", matches[1])
			currentName = matches[1]
			currentInt = Interface{
				Description: "",
				MacAddress:  "",
				OperStatus:  "down",
				AdminStatus: "down",
				RxBytes:     0,
				TxBytes:     0,
				RxPackets:   0,
				TxPackets:   0,
				RxUnicast:   0,
				TxUnicast:   0,
				RxBcast:     0,
				TxBcast:     0,
				RxMcast:     0,
				TxMcast:     0,
				RxDrops:     0,
				TxDrops:     0,
				RxErrors:    0,
				TxErrors:    0,
			}
			if strings.ToLower(matches[2]) == "up" {
				log.Debugf("OperStatus: %+v", matches[2])
				currentInt.OperStatus = "up"
			}
			continue
		}

		if matches := adminStateRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("AdminStatus: %+v", matches[1])
			if strings.ToLower(matches[1]) == "up" {
				currentInt.AdminStatus = "up"
			}
			continue
		}

		if matches := descRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("Description: %+v", matches[1])
			currentInt.Description = matches[1]
			continue
		}

		if matches := macRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("MacAddress: %+v", matches[1])
			currentInt.MacAddress = util.StandardizeMacAddr(matches[1])
			continue
		}

		if matches := packetsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxPackets: %+v", matches[1])
			log.Debugf("TxPackets: %+v", matches[2])
			currentInt.RxPackets += util.Str2float64(matches[1])
			currentInt.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := l3packetsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("L3RxPackets: %+v", matches[1])
			log.Debugf("L3TxPackets: %+v", matches[2])
			currentInt.RxPackets += util.Str2float64(matches[1])
			currentInt.TxPackets += util.Str2float64(matches[2])
			continue
		}

		if matches := unicastRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxUnicast: %+v", matches[1])
			log.Debugf("TxUnicast: %+v", matches[2])
			currentInt.RxUnicast += util.Str2float64(matches[1])
			currentInt.TxUnicast += util.Str2float64(matches[2])
			continue
		}

		if matches := McastRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxMcast: %+v", matches[1])
			log.Debugf("TxMcast: %+v", matches[2])
			currentInt.RxMcast += util.Str2float64(matches[1])
			currentInt.TxMcast += util.Str2float64(matches[2])
			continue
		}

		if matches := BcastRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxBcast: %+v", matches[1])
			log.Debugf("TxBcast: %+v", matches[2])
			currentInt.RxBcast += util.Str2float64(matches[1])
			currentInt.TxBcast += util.Str2float64(matches[2])
			continue
		}

		if matches := bytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxBytes: %+v", matches[1])
			log.Debugf("TxBytes: %+v", matches[2])
			currentInt.RxBytes += util.Str2float64(matches[1])
			currentInt.TxBytes += util.Str2float64(matches[2])
			continue
		}

		if matches := l3bytesRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("L3RxBytes: %+v", matches[1])
			log.Debugf("L3TxBytes: %+v", matches[2])
			currentInt.RxBytes += util.Str2float64(matches[1])
			currentInt.TxBytes += util.Str2float64(matches[2])
			continue
		}

		if matches := dropsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxDrops: %+v", matches[1])
			log.Debugf("TxDrops: %+v", matches[2])
			currentInt.RxDrops += util.Str2float64(matches[1])
			currentInt.TxDrops += util.Str2float64(matches[2])
			continue
		}

		if matches := errorsRegexp.FindStringSubmatch(line); matches != nil {
			log.Debugf("RxErrors: %+v", matches[1])
			log.Debugf("TxErrors: %+v", matches[2])
			currentInt.RxErrors += util.Str2float64(matches[1])
			currentInt.TxErrors += util.Str2float64(matches[1])
			continue
		}

	}

	interfaces[currentName] = currentInt

	return interfaces, nil
}
