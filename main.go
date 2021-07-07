package main

import (
	"bytes"
        "flag"
	"fmt"
	"io/ioutil"
	"os"
//	"os/signal"
//	"strings"
//	"time"

        "github.com/yankiwi/aruba_exporter/config"
        "github.com/yankiwi/aruba_exporter/connector"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const version string = "0.0.1"

var (
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("web.listen-address", ":9326", "Address on which to expose metrics and web interface.")
	metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sshHosts           = flag.String("ssh.targets", "", "Hosts to scrape")
	sshUsername        = flag.String("ssh.user", "junos_exporter", "Username to use when connecting to junos devices using ssh")
	sshKeyFile         = flag.String("ssh.keyfile", "", "Public key file to use when connecting to junos devices using ssh")
	sshPassword        = flag.String("ssh.password", "", "Password to use when connecting to junos devices using ssh")
	sshTimeout         = flag.Int("ssh.timeout", 5, "Timeout to use for SSH connection")
	sshBatchSize       = flag.Int("ssh.batch-size", 10000, "The SSH response batch size")
	debug              = flag.Bool("debug", false, "Show verbose debug output in log")
	configFile         = flag.String("config.file", "", "Path to config file")
	devices            []*connector.Device
	cfg                *config.Config
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: aruba_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		log.Fatalf("could not initialize exporter. %v", err)
	}

}

func initialize() error {
	c, err := loadConfig()
	if err != nil {
		return err
	}

	devices, err = devicesForConfig(c)
	if err != nil {
		return err
	}
	cfg = c

	return nil
}

func loadConfig() (*config.Config, error) {
	if len(*configFile) == 0 {
		log.Infoln("Loading config flags")
		return loadConfigFromFlags(), nil
	}

	log.Infoln("Loading config from", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func loadConfigFromFlags() *config.Config {
	c := config.New()

	c.Debug = *debug
	c.Timeout = *sshTimeout
	c.BatchSize = *sshBatchSize
	c.Username = *sshUsername
	c.Password = *sshPassword
	c.KeyFile = *sshKeyFile
	c.DevicesFromTargets(*sshHosts)
	log.Infoln(c)

	f := c.Features
	log.Infoln(f)

	return c
}

func printVersion() {
	fmt.Println("aruba_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Patrick Ryon")
	fmt.Println("Metric exporter for Aruba switches, controllers and instant APs")
}

