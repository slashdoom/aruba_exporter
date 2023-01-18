package wireless

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/slashdoom/aruba_exporter/rpc"
	"github.com/slashdoom/aruba_exporter/util"
	
	log "github.com/sirupsen/logrus"
)

// ParseAccessPoints parases cli output and tries to find AP info
func (c *wirelessCollector) ParseAccessPoints(ostype string, output string) (map[string]WirelessAccessPoint, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Tracef("output: %s\n", output)

	aps := make(map[string]WirelessAccessPoint)

	lines := strings.Split(output, "\n")
	
	if ostype == rpc.ArubaController {
		return aps, nil
	}
	if ostype == rpc.ArubaInstant {
		ConductorIPRegexp, _ := regexp.Compile(`^^Conductor IP Address\s+(\*?)\:`)
		APIPRegexp, _ := regexp.Compile(`^IP Address\s+\:(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s*$`)
		APNameRegexp, _ := regexp.Compile(`^!~~~NO_MATCHES~~~!$`)

		ap := WirelessAccessPoint{}
		currentAPIP:= ""
		
		for _, line := range lines {
			
			log.Tracef("line: %+v", line)
			
			if matches := ConductorIPRegexp.FindStringSubmatch(line); matches != nil {
				ap = WirelessAccessPoint{Controller: false}
				currentAPIP = ""
				if matches[1] == "*" {
					log.Debugf("controller: %+v\n", true)
					ap.Controller = true
				}
				continue
			}
			if matches := APIPRegexp.FindStringSubmatch(line); matches != nil {
				currentAPIP = matches[1]
				log.Debugf("AP IP: %+v\n", currentAPIP)
				APNameRegexp, _ = regexp.Compile(fmt.Sprintf(`^.+\s+%v\s+(.+?)\s+(\d+)\s+`, currentAPIP))
				log.Debugf("AP Name Regexp: %+v\n", fmt.Sprintf(`^.+\s+%v\s+(.+?)\s+(\d+)\s+`, currentAPIP))
				continue
			}
			if matches := APNameRegexp.FindStringSubmatch(line); matches != nil {
				ap.Name = matches[1]
				log.Debugf("AP Name: %+v\n", ap.Name)
				ap.Clients = util.Str2float64(matches[2])
				log.Debugf("AP Clients: %+v\n", ap.Clients)
				ap.Up = true
				aps[currentAPIP] = ap
				break
			}
		}
		log.Debugf("AP Data: %+v\n", aps)
		return aps, nil
	}
	return make(map[string]WirelessAccessPoint), errors.New("AP info not found")
}

// ParseChannels parses cli output and tries to find AP channel stats
func (c *wirelessCollector) ParseChannels(ostype string, output string) (map[string]WirelessChannel, map[string]WirelessRadio, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Tracef("output: %s\n", output)
	
	channels := make(map[string]WirelessChannel)
	radios := make(map[string]WirelessRadio)

	lines := strings.Split(output, "\n")
	currentInt := ""
	
	if ostype == rpc.ArubaController {
		return channels, radios, nil
	}
	if ostype == rpc.ArubaInstant {
		nameRegexp, _ := regexp.Compile(`^name\:(.+)$`)
		channelRegexp, _ := regexp.Compile(`^(\d\.?\d?)GHz\s+(\d+)\s+\d+\s+\d+\s+\d+\s+(\d+)\s+(\d+)\/\d+\/\d+\/\d+\/(\d+)\s+\d+\/\d+\((\d+)\)\s+\d+\/\d+\/\/\d+\/\d+\((\d+)\)$\s*$`)
		interfaceRegexp, _ := regexp.Compile(`^Interface Name\s+:wifi(\d)\s*$`)
		bandRegexp, _ := regexp.Compile(`^Phy-Type\s+:(\d\.?\d?)GHz\s*$`)
		assignmentRegexp, _ := regexp.Compile(`^Current ARM Assignment\s+:(\d+)\+?\/(\d+\.?\d?)\s*$`)
		apName := ""
		for _, line := range lines {
			log.Debugf("line: %+v", line)
			if matches := nameRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("AP name: %v\n", matches[1])
				apName = matches[1]
				continue
			}
			if matches := channelRegexp.FindStringSubmatch(line); matches != nil {
				channel := WirelessChannel{
					AccessPoint: apName,
					Band: util.Str2float64(matches[1]),
					NoiseFloor: util.Str2float64(matches[3]),
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

			if matches := interfaceRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("arm radio: %+v\n", matches[1])
				currentInt = matches[1]
				radios[currentInt] = WirelessRadio{}
				continue
			}
			if matches := bandRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("arm radio %+v band: %v\n", currentInt, matches[1])
				if radio, ok := radios[currentInt]; ok {
					radio.Band = util.Str2float64(matches[1])
					radios[currentInt] = radio
				}
				continue
			}
			if matches := assignmentRegexp.FindStringSubmatch(line); matches != nil {
				log.Debugf("arm radio %+v channel: %v\n", currentInt, matches[1])
				log.Debugf("arm radio %+v power: %v\n", currentInt, matches[2])
				if radio, ok := radios[currentInt]; ok {
					radio.Channel = util.Str2float64(matches[1])
					radio.Power = util.Str2float64(matches[2])
					radio.ChUtil = channels[matches[1]].ChUtil
					radio.ChQual = channels[matches[1]].ChQual
					radio.NoiseFloor = channels[matches[1]].NoiseFloor
					radios[currentInt] = radio
				}
				continue
			}
		}

		return channels, radios, nil
	}
                                   
	return make(map[string]WirelessChannel), make(map[string]WirelessRadio), errors.New("Channel info not found")
}

// ParseRadios parses cli output and tries to find AP radio stats
func (c *wirelessCollector) ParseRadios(ostype string, radios map[string]WirelessRadio, output string) (map[string]WirelessRadio, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Tracef("output: %s\n", output)

	return make(map[string]WirelessRadio), nil
}