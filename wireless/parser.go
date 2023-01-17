package wireless

import (
	"errors"
//	"fmt"
	"regexp"
	"strings"

	"github.com/slashdoom/aruba_exporter/rpc"
	"github.com/slashdoom/aruba_exporter/util"
	
	log "github.com/sirupsen/logrus"
)

// ParseChannels parses cli output and tries to find AP channel stats
func (c *wirelessCollector) ParseChannels(ostype string, output string) (map[string]WirelessChannel, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Debugf("output: %s\n", output)
	
	channels := make(map[string]WirelessChannel)
	lines := strings.Split(output, "\n")
	
	if ostype == rpc.ArubaController {
		return channels, nil
	}
	if ostype == rpc.ArubaInstant {
		for _, line := range lines {
			channelRegexp, _ := regexp.Compile(`^(\d\.?\d?)GHz\s+(\d+)\s+\d+\s+\d+\s+\d+\s+(\d+)\s+(\d+)\/\d+\/\d+\/\d+\/(\d+)\s+\d+\/\d+\((\d+)\)\s+\d+\/\d+\/\/\d+\/\d+\((\d+)\)$`)

			log.Tracef("line: %+v", line)
			if matches := channelRegexp.FindStringSubmatch(line); matches != nil {
				channel := WirelessChannel{
					Band: util.Str2float64(matches[1]),
					Noise: util.Str2float64(matches[3]),
					ChUtil: util.Str2float64(matches[4]),
					ChQual: util.Str2float64(matches[5]),
					CovrIndex: util.Str2float64(matches[6]),
					IntfIndex: util.Str2float64(matches[7]),
				}
				channels[matches[2]] = channel
				log.Debugf("channel name: %+v\n", matches[2])
				log.Debugf("channel data: %+v\n", channel)
				continue
			}
		}

		return channels, nil
	}
                                   
	return make(map[string]WirelessChannel), errors.New("Channels info not found")
}