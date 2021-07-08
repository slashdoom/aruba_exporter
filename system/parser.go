package system

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *systemCollector) ParseVersion(ostype string, output string) (SystemVersion, error) {
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController {
		return SystemVersion{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.ArubaInstant], _ = regexp.Compile(`^.*, Version (.*)$`)
	versionRegexp[rpc.ArubaController], _ = regexp.Compile(`^.*, Version (.*)$`)

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
	}
	if ostype == rpc.ArubaInstant {
		totalMemRegexp, _ := regexp.Compile(`^.*MemTotal:\s*(\d+) kB.*$`)
		freeMemRegexp, _ := regexp.Compile(`^.*MemFree:\s*(\d+) kB.*$`)
		availMemRegexp, _ := regexp.Compile(`^.*MemAvailable:\s*(\d+) kB.*$`)
		
		for _, line := range lines {
			totalMatches := totalMemRegexp.FindStringSubmatch(line)
			freeMatches := freeMemRegexp.FindStringSubmatch(line)
			availMatches := availMemRegexp.FindStringSubmatch(line)
			if totalMatches == nil && freeMatches == nil && availMatches == nil {
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
	}
	                                   
	return items, nil
}

// ParseCPU parses cli output and tries to find current CPU utilization
func (c *systemCollector) ParseCPU(ostype string, output string) (SystemCPU, error) {
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController {
		return SystemCPU{}, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*CPU utilization for five seconds: (\d+)%\/(\d+)%; one minute: (\d+)%; five minutes: (\d+)%.*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return SystemCPU{
			FiveSeconds: util.Str2float64(matches[1]),
			Interrupts:  util.Str2float64(matches[2]),
			OneMinute:   util.Str2float64(matches[3]),
			FiveMinutes: util.Str2float64(matches[4]),
		}, nil
	}
	return SystemCPU{}, errors.New("Version string not found")
}
