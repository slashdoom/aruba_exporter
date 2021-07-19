package system

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"
	
	"github.com/prometheus/common/log"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *systemCollector) ParseVersion(ostype string, output string) (SystemVersion, error) {
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController {
		return SystemVersion{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.ArubaInstant], _ = regexp.Compile(`^.*, Version (.*)$`)
	versionRegexp[rpc.ArubaController], _ = regexp.Compile(`^.*, Version (.*)$`)
	versionRegexp[rpc.ArubaSwitch], _ = regexp.Compile(`^\s*[A-Z]{2}.(.*)$`)
	versionRegexp[rpc.ArubaCXSwitch], _ = regexp.Compile(`^.*Version\s*:\s*[A-Z]{2}.(.*)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := versionRegexp[ostype].FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return SystemVersion{Version: ostype + "-" + matches[1]}, nil
	}
	return SystemVersion{}, errors.New("Version string not found")
}

// ParseMemory parses cli output and tries to find current memory usage
func (c *systemCollector) ParseMemory(ostype string, output string) ([]SystemMemory, error) {
	log.Infof("OS: %s\n", ostype)
	log.Infof("output: %s\n", output)
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController {
		return nil, errors.New("'show memory' is not implemented for " + ostype)
	}
	
	items := []SystemMemory{}
	lines := strings.Split(output, "\n")
	
	if ostype == rpc.ArubaController {
		memoryRegexp, _ := regexp.Compile(`^.*Memory (Kb): total:\s*(\d+), used:\s*(\d+), free:\s*(\d+)\s*$`)
		
		for _, line := range lines {
			matches := memoryRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemMemory{
				Type:  matches[1],
				Total: util.Str2float64(matches[2]),
				Used:  util.Str2float64(matches[3]),
				Free:  util.Str2float64(matches[4]),
			}
			items = append(items, item)
		}
		return items, nil
	}
	if ostype == rpc.ArubaInstant {
		totalMemRegexp, _ := regexp.Compile(`^.*MemTotal:\s*(\d+) kB.*$`)
		freeMemRegexp, _ := regexp.Compile(`^.*MemFree:\s*(\d+) kB.*$`)
		availMemRegexp, _ := regexp.Compile(`^.*MemAvailable:\s*(\d+) kB.*$`)
		
		for _, line := range lines {
			log.Infof("line: %s\n", line)
			totalMatches := totalMemRegexp.FindStringSubmatch(line)
			freeMatches := freeMemRegexp.FindStringSubmatch(line)
			availMatches := availMemRegexp.FindStringSubmatch(line)
			if totalMatches == nil || freeMatches == nil || availMatches == nil {
				continue
			}
			totalMem := util.Str2float64(totalMatches[2])
			freeMem := util.Str2float64(freeMatches[2])
			usedMem := totalMem - util.Str2float64(availMatches[2])
			
			item := SystemMemory{
				Type:  fmt.Sprintf("Memory (Kb): total: %d, used: %d, free: %d", totalMem, usedMem, freeMem),
				Total: totalMem,
				Used:  usedMem,
				Free:  freeMem,
			}
			items = append(items, item)
		}
		return items, nil
	}
	                                   
	return []SystemMemory{}, errors.New("Memory string not found")
}

// ParseCPU parses cli output and tries to find current CPU utilization
func (c *systemCollector) ParseCPU(ostype string, output string) ([]SystemCPU, error) {
	log.Infof("OS: %s\n", ostype)
	log.Infof("output: %s\n", output)
	if ostype != rpc.ArubaInstant {
		return nil, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	items := []SystemCPU{}
	lines := strings.Split(output, "\n")

	if ostype == rpc.ArubaInstant {
		cpuRegexp, _ := regexp.Compile(`^\s*(.+): user\s*(\d+)% nice\s*(\d+)% system\s*(\d+)% idle\s*(\d+)% io\s*(\d+)% irq\s*(\d+)% softirq\s*(\d+)%.*$`)

		for _, line := range lines {
			log.Infof("line: %s\n", line)
			matches := cpuRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			return SystemCPU{
				Type: matches[1],
				Used: util.Str2float64(matches[2])+util.Str2float64(matches[3])+util.Str2float64(matches[4]),
				Idle: util.Str2float64(matches[5]),
			}, nil
		}
	}

	return []SystemCPU{}, errors.New("CPU string not found")
}
