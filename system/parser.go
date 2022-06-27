package system

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"
	
	log "github.com/sirupsen/logrus"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *systemCollector) ParseVersion(ostype string, output string) (SystemVersion, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Debugf("output: %s\n", output)
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController && ostype != rpc.ArubaSwitch && ostype != rpc.ArubaCXSwitch {
		return SystemVersion{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.ArubaInstant], _ = regexp.Compile(`^.*, Version (.*)$`)
	versionRegexp[rpc.ArubaController], _ = regexp.Compile(`^.*, Version (.*)$`)
	versionRegexp[rpc.ArubaSwitch], _ = regexp.Compile(`^\s*([A-Z]{2}\..*)$`)
	versionRegexp[rpc.ArubaCXSwitch], _ = regexp.Compile(`^Version\s*:\s*([A-Z]{2}\.\d{2}\.\d{2}\.\d{4})\s*$`)

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
	log.Debugf("OS: %s\n", ostype)
	log.Debugf("output: %s\n", output)
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController && ostype != rpc.ArubaSwitch && ostype != rpc.ArubaCXSwitch {
		return nil, errors.New("'show memory' is not implemented for " + ostype)
	}
	
	items := []SystemMemory{}
	lines := strings.Split(output, "\n")
	
	if ostype == rpc.ArubaController {
		memoryRegexp, _ := regexp.Compile(`^.*Memory \(Kb\): total:\s*(\d+), used:\s*(\d+), free:\s*(\d+)\s*$`)
		
		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matches := memoryRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemMemory{
				Type:  "system",
				Total: util.Str2float64(matches[1]),
				Used:  util.Str2float64(matches[2]),
				Free:  util.Str2float64(matches[3]),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}
	if ostype == rpc.ArubaInstant {
		totalMemRegexp, _ := regexp.Compile(`^.*MemTotal:\s*(\d+) kB.*$`)
		freeMemRegexp, _ := regexp.Compile(`^.*MemFree:\s*(\d+) kB.*$`)
		availMemRegexp, _ := regexp.Compile(`^.*MemAvailable:\s*(\d+) kB.*$`)
		var (
			totalMem SystemValue
			freeMem SystemValue
			usedMem SystemValue
		)
		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			totalMatches := totalMemRegexp.FindStringSubmatch(line)
			freeMatches := freeMemRegexp.FindStringSubmatch(line)
			availMatches := availMemRegexp.FindStringSubmatch(line)

			if !totalMem.isSet && totalMatches != nil {
				log.Debugf("totalMatches: %+v", totalMatches)
				totalMem.isSet = true
				totalMem.Value = util.Str2float64(totalMatches[1])
			}
			if !freeMem.isSet && freeMatches != nil {
				log.Debugf("freeMatches: %+v", freeMatches)
				freeMem.isSet = true
				freeMem.Value = util.Str2float64(freeMatches[1])
			}
			if !usedMem.isSet && availMatches != nil {
				log.Debugf("availMatches: %+v", availMatches)
				usedMem.isSet = true
				usedMem.Value = util.Str2float64(availMatches[1])
			}

			if !totalMem.isSet || !freeMem.isSet || !usedMem.isSet {
				continue
			}
			item := SystemMemory{
				Type: fmt.Sprintf("system"),
				Total: totalMem.Value,
				Used: usedMem.Value,
				Free: (totalMem.Value - usedMem.Value),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
			break
		}
		return items, nil
	}
	if ostype == rpc.ArubaSwitch {
		totalMemRegexp, _ := regexp.Compile(`System Total Memory\(bytes\):\s*(\d+)`)
		usedMemRegexp, _ := regexp.Compile(`Total Used Memory\(bytes\):\s*(\d+)`)
		var (
			totalMem SystemValue
			usedMem SystemValue
		)
		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			totalMatches := totalMemRegexp.FindStringSubmatch(line)
			usedMatches := usedMemRegexp.FindStringSubmatch(line)

			if !totalMem.isSet && totalMatches != nil {
				log.Debugf("totalMatches: %+v", totalMatches)
				totalMem.isSet = true
				totalMem.Value = util.Str2float64(totalMatches[1])
			}
			if !usedMem.isSet && usedMatches != nil {
				log.Debugf("usedMatches: %+v", usedMatches)
				usedMem.isSet = true
				usedMem.Value = util.Str2float64(usedMatches[1])
			}

			if !totalMem.isSet || !usedMem.isSet {
				continue
			}
			item := SystemMemory{
				Type: fmt.Sprintf("system"),
				Total: math.RoundToEven(totalMem.Value/1000),
				Used: math.RoundToEven(usedMem.Value/1000),
				Free: math.RoundToEven((totalMem.Value - usedMem.Value)/1000),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
			break
		}
		return items, nil
	}
	if ostype == rpc.ArubaCXSwitch {
		memoryRegexp, _ := regexp.Compile(`^MiB Mem\s*:\s*(\d+\.\d+) total,\s*(\d+\.\d+) free,\s*(\d+\.\d+) used,\s*(\d+\.\d+) buff/cache\s*$`)
		swapRegexp, _ := regexp.Compile(`^MiB Swap\s*:\s*(\d+\.\d+) total,\s*(\d+\.\d+) free,\s*(\d+\.\d+) used.\s*(\d+\.\d+) avail Mem\s*$`)

		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matchesMem := memoryRegexp.FindStringSubmatch(line)
			matchesSwap := swapRegexp.FindStringSubmatch(line)
			if matchesMem == nil && matchesSwap == nil {
				continue
			}
			item := SystemMemory{}
			if matchesMem != nil {
				item = SystemMemory{
					Type:  "system",
					Total: util.Str2float64(matchesMem[1])*1000,
					Used:  util.Str2float64(matchesMem[3])*1000,
					Free:  util.Str2float64(matchesMem[2])*1000,
				}
			}
			if matchesSwap != nil {
				item = SystemMemory{
					Type:  "swap",
					Total: util.Str2float64(matchesSwap[1])*1000,
					Used:  util.Str2float64(matchesSwap[3])*1000,
					Free:  util.Str2float64(matchesSwap[2])*1000,
				}
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}
	                                   
	return []SystemMemory{}, errors.New("Memory string not found")
}

// ParseCPU parses cli output and tries to find current CPU utilization
func (c *systemCollector) ParseCPU(ostype string, output string) ([]SystemCPU, error) {
	log.Debugf("OS: %s\n", ostype)
	log.Debugf("output: %s\n", output)
	if ostype != rpc.ArubaInstant && ostype != rpc.ArubaController && ostype != rpc.ArubaSwitch && ostype != rpc.ArubaCXSwitch {
		return nil, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	items := []SystemCPU{}
	lines := strings.Split(output, "\n")

	if ostype == rpc.ArubaController {
		cpuRegexp, _ := regexp.Compile(`^.*user\s*(\d+\.?\d*)%, system\s*(\d+\.?\d*)%, idle\s*(\d+\.?\d*)%.*$`)

		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matches := cpuRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemCPU{
				Type: "total",
				Used: (util.Str2float64(matches[1])+util.Str2float64(matches[2])),
				Idle: util.Str2float64(matches[3]),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}
	if ostype == rpc.ArubaInstant {
		cpuRegexp, _ := regexp.Compile(`^\s*(.+): user\s*(\d+)% nice\s*(\d+)% system\s*(\d+)% idle\s*(\d+)% io\s*(\d+)% irq\s*(\d+)% softirq\s*(\d+)%.*$`)                      

		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matches := cpuRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemCPU{
				Type: matches[1],
				Used: (util.Str2float64(matches[2])+util.Str2float64(matches[3])+util.Str2float64(matches[4])),
				Idle: util.Str2float64(matches[5]),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}
	if ostype == rpc.ArubaSwitch {
		cpuRegexp, _ := regexp.Compile(`^(\d+) percent busy, from \d+ sec ago$`)

		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matches := cpuRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemCPU{
				Type: "total",
				Used: util.Str2float64(matches[1]),
				Idle: 100-util.Str2float64(matches[1]),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}
	if ostype == rpc.ArubaCXSwitch {
		cpuRegexp, _ := regexp.Compile(`^CPU Util \(%\)\s*:\s*(\d+)\s*$`)

		for _, line := range lines {
			log.Debugf("line: %s\n", line)
			matches := cpuRegexp.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			item := SystemCPU{
				Type: "total",
				Used: util.Str2float64(matches[1]),
				Idle: 100-util.Str2float64(matches[1]),
			}
			log.Debugf("item: %+v\n", item)
			items = append(items, item)
		}
		return items, nil
	}

	return []SystemCPU{}, errors.New("CPU string not found")
}
