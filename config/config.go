package config

import (
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents the configuration for the exporter
type Config struct {
	Level         string          `yaml:"level,omitempty"`
	LegacyCiphers bool            `yaml:"legacy_ciphers,omitempty"`
	Timeout       int             `yaml:"timeout,omitempty"`
	BatchSize     int             `yaml:"batch_size,omitempty"`
	Username      string          `yaml:"username,omitempty"`
	Password      string          `yaml:"password,omitempty"`
	KeyFile       string          `yaml:"key_file,omitempty"`
	Devices       []*DeviceConfig `yaml:"devices,omitempty"`
	Features      *FeatureConfig  `yaml:"features,omitempty"`
}

// DeviceConfig is the config representation of 1 device
type DeviceConfig struct {
	Host          string         `yaml:"host"`
	Username      *string        `yaml:"username,omitempty"`
	Password      *string        `yaml:"password,omitempty"`
	KeyFile       *string        `yaml:"key_file,omitempty"`
	LegacyCiphers *bool          `yaml:"legacy_ciphers,omitempty"`
	Timeout       *int           `yaml:"timeout,omitempty"`
	BatchSize     *int           `yaml:"batch_size,omitempty"`
	Features      *FeatureConfig `yaml:"features,omitempty"`
}

// FeatureConfig is the list of collectors enabled or disabled
type FeatureConfig struct {
	BGP         *bool `yaml:"bgp,omitempty"`
	Environment *bool `yaml:"environment,omitempty"`
	Interfaces  *bool `yaml:"interfaces,omitempty"`
	Optics      *bool `yaml:"optics,omitempty"`
	System      *bool `yaml:"system,omitempty"`
	Wireless	*bool `yaml:"wireless,omitempty"`
}

// New creates a new config
func New() *Config {
	c := &Config{
		Features: &FeatureConfig{},
	}
	c.setDefaultValues()

	return c
}

// Load loads a config from reader
func Load(reader io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	c := New()
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) setDefaultValues() {
	c.Level = "info"
	c.LegacyCiphers = false
	c.Timeout = 5
	c.BatchSize = 10000

	f := c.Features
	bgp := true
	f.BGP = &bgp
	environment := true
	f.Environment = &environment
	interfaces := true
	f.Interfaces = &interfaces
	optics := true
	f.Optics = &optics
	system := true
	f.System = &system
	wireless := true
	f.Wireless = &wireless
}

// DevicesFromTargets creates devices configs from targets list
func (c *Config) DevicesFromTargets(sshHosts string) {
	targets := strings.Split(sshHosts, ",")

	c.Devices = make([]*DeviceConfig, len(targets))
	for i, target := range targets {
		c.Devices[i] = &DeviceConfig{
			Host: target,
		}
	}
}

// FeaturesForDevice gets the feature set configured for a device
func (c *Config) FeaturesForDevice(host string) *FeatureConfig {
	d := c.findDeviceConfig(host)

	if d != nil && d.Features != nil {
		return d.Features
	}

	return c.Features
}

func (c *Config) findDeviceConfig(host string) *DeviceConfig {
	for _, dc := range c.Devices {
		if dc.Host == host {
			return dc
		}
	}

	return nil
}
