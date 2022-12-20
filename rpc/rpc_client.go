package rpc

import (
	"errors"
	"strings"

	"github.com/yankiwi/aruba_exporter/connector"
	
	log "github.com/sirupsen/logrus"
)

const (
	ArubaInstant string = "ArubaInstant"
	ArubaController string = "ArubaController"
	ArubaSwitch string = "ArubaSwitch"
	ArubaCXSwitch string = "ArubaCXSwitch"
)

// Client sends commands to a Aruba device
type Client struct {
	conn   *connector.SSHConnection
	Level  string
	OSType string
}

// NewClient creates a new client connection
func NewClient(ssh *connector.SSHConnection, level string) *Client {
	rpc := &Client{conn: ssh, Level: level}

	return rpc
}

// Identify tries to identify the OS running on a Aruba device
func (c *Client) Identify() error {
	output, err := c.RunCommand([]string{"show version"})
	if err != nil {
		return err
	}
	
	log.Debugf("show version output: %s\n", output)
	
	switch {
	case strings.Contains(output, "ArubaOS (MODEL: Aruba"):
		c.OSType = ArubaController
	case strings.Contains(output, "ArubaOS (MODEL: "):
		c.OSType = ArubaInstant
	case strings.Contains(output, "/ws/swbuild"):
		c.OSType = ArubaSwitch
	case strings.Contains(output, "ArubaOS-CX"):
		c.OSType = ArubaCXSwitch
	default:
		return errors.New("Unknown OS")
	}

	log.Infof("Host %s identified as: %s\n", c.conn.Host, c.OSType)

	return nil
}

// RunCommand runs a command on a Aruba device
func (c *Client) RunCommand(cmds []string) (string, error) {

	output, err := c.conn.RunCommand(cmds)
	if err != nil {
		log.Errorln(err.Error())
		return "", err
	}

	return output, nil
}
