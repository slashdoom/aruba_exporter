package connector

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/slashdoom/aruba_exporter/config"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	log "github.com/sirupsen/logrus"
)

// SSHConnection encapsulates the connection to the device
type SSHConnection struct {
	client       *ssh.Client
	Host         string
	stdin        io.WriteCloser
	stdout       io.Reader
	session      *ssh.Session
	batchSize    int
	clientConfig *ssh.ClientConfig
}

type result struct {
	output string
	err    error
}

// NewSSSHConnection connects to device
func NewSSSHConnection(device *Device, cfg *config.Config) (*SSHConnection, error) {
	deviceConfig := device.DeviceConfig

	legacyCiphers := cfg.LegacyCiphers
	if deviceConfig.LegacyCiphers != nil {
		legacyCiphers = *deviceConfig.LegacyCiphers
	}

	batchSize := cfg.BatchSize
	if deviceConfig.BatchSize != nil {
		batchSize = *deviceConfig.BatchSize
	}

	timeout := cfg.Timeout
	if deviceConfig.Timeout != nil {
		timeout = *deviceConfig.Timeout
	}

	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}
	if legacyCiphers {
		sshConfig.SetDefaults()
		sshConfig.Ciphers = append(sshConfig.Ciphers, "aes128-cbc", "3des-cbc")
	}

	device.Auth(sshConfig)

	c := &SSHConnection{
		Host:         device.Host + ":" + device.Port,
		batchSize:    batchSize,
		clientConfig: sshConfig,
	}

	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Connect connects to the device
func (c *SSHConnection) Connect() error {
	var (
		err error
		output string
	)
	c.client, err = ssh.Dial("tcp", c.Host, c.clientConfig)
	if err != nil {
		return err
	}

	session, err := c.client.NewSession()
	if err != nil {
		c.client.Conn.Close()
		return err
	}
	c.stdin, _ = session.StdinPipe()
	c.stdout, _ = session.StdoutPipe()
	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
		ssh.ECHOCTL: 0,
		ssh.OCRNL: 0,
	}
	session.RequestPty("vt100", 0, 2000, modes)
	session.Shell()
	c.session = session

	c.BlindSend([]string{"", ""})
	output, err = c.RunCommand([]string{""})
	output, err = c.RunCommand([]string{"no page"})
	log.Traceln(output, err)
	output, err = c.RunCommand([]string{"no paging"})
	log.Traceln(output, err)
	output, err = c.RunCommand([]string{""})

	return nil
}

// RunCommand runs a command or commands against the device
func (c *SSHConnection) RunCommand(cmds []string) (string, error) {
	buf := bufio.NewReader(c.stdout)
	for i := 0; i < len(cmds); i++ {
		log.Debugf("Running command on %s: %s\n", c.Host, cmds[i])
		io.WriteString(c.stdin, fmt.Sprintf("%s", cmds[i])+"\n")
		time.Sleep(1 * time.Second)
	}

	outputChan := make(chan result)
	go func() {
		c.readln(outputChan, cmds[0], buf)
	}()
	select {
	case res := <-outputChan:
		return res.output, res.err
	case <-time.After(c.clientConfig.Timeout):
		return "", errors.New("Timeout reached")
	}
}

// BlindSend sends commands to a device and doesn't wait for output
func (c *SSHConnection) BlindSend(cmds []string) {
	for i := 0; i < len(cmds); i++ {
		log.Debugf("Blind sending command on %s: %s\n", c.Host, cmds[i])
		io.WriteString(c.stdin, fmt.Sprintf("%s", cmds[i])+"\n")
	}
	time.Sleep(1 * time.Second)
}

// Close closes connection
func (c *SSHConnection) Close() {
	if c.client.Conn == nil {
		return
	}
	c.client.Conn.Close()
	if c.session != nil {
		c.session.Close()
	}
}

func loadPrivateKey(r io.Reader) (ssh.AuthMethod, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not read from reader")
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse private key")
	}

	return ssh.PublicKeys(key), nil
}

func (c *SSHConnection) readln(ch chan result, cmd string, r io.Reader) {
	endPrompt := regexp.MustCompile(`(?m).+#\s*?$`)
	escSequence := regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
	buf := make([]byte, c.batchSize)
	loadStr := ""
	for {
		n, err := r.Read(buf)
		if err != nil {
			ch <- result{output: "", err: err}
		}
		cleanStr := escSequence.ReplaceAllString(string(buf[:n]), "")
		loadStr += cleanStr
		log.Debugln(loadStr)
		if strings.Contains(loadStr, cmd) {
			log.Debugln("command match")
		    if endPrompt.MatchString(loadStr) {
				log.Debugln("prompt match")
			    break
			}
		}
	}
	loadStr = strings.Replace(loadStr, "\r", "", -1)
	ch <- result{output: loadStr, err: nil}
}

