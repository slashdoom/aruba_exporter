# aruba_exporter
Prometheus exporter for metrics from Aruba devices including ArubaSwitchOS, ArubaOS-CX, ArubaOS (Instant AP and controllers/gateways).

This exporter is in development.  The goal to get a basic set of metrics working on the four major devices types via SSH.  Additional metrics will be developed as required or requested.

The basic structure is based on:

https://github.com/czerwonk/junos_exporter  
...and...  
https://github.com/lwlcom/cisco_exporter  

# Flags
Name     | Description | Default
---------|-------------|---------
version | Print version information. |
web.listen-address | Address on which to expose metrics and web interface. | :9909
web.telemetry-path | Path under which to expose metrics. | /metrics
ssh.targets | Comma seperated list of hosts to scrape |
ssh.user | Username to use when connecting to devices using ssh. | aruba_exporter
ssh.keyfile | Public key file to use when connecting to devices using ssh. |
ssh.password | Password to use when connecting to devices using ssh. |
ssh.timeout | Timeout in seconds to use for SSH connection. | 5
ssh.batch-size | The SSH response batch size. | 10000
level | Set logging verbose level. | info
config.file | Path to config file. |

# Metrics
All metrics are enabled by default. To disable something pass a flag `--<name>.enabled=false`, where `<name>` is the name of the metric.

Name     | Description | SwitchOS | OS-CX | InstantAP | Controller |
---------|-------------|----------|-------|-----------|------------|
system | System metrics (version, CPU (% used/idle), memory (total/used/free), uptime) | X | X | X | X |
environment | Environment metrics (temperatures, state of power supply) | - | - | - | - |
interfaces | Interfaces metrics (transmitted/received: bytes/packets/errors/drops, admin/oper state) | X | X | X | X |
optics | Optical signals metrics (tx/rx) | - | - | - | - |
routes | Router metrics (total, static, dynamic, connected) | - | - | N/A | - |
wifi | Wi-Fi metrics (clients, aps, wlans) | N/A | N/A | - | - |

# Install
```bash
go get -u github.com/slashdoom/aruba_exporter
```

# Usage

## Binary
```bash
./aruba_exporter -ssh.targets="host1.example.com,host2.example.com:2233,172.16.0.1" -ssh.keyfile=aruba_exporter

./aruba_exporter -ssh.targets="host1.example.com,host2.example.com:2233,172.16.0.1" -ssh.password=password

./aruba_exporter -config.file=config.yml
```

## Config file
The exporter can be configured with a YAML based config file:

```yaml
---
level: debug
timeout: 60
batch_size: 10000
username: default-username
password: default-password
key_file: /path/to/key

devices:
  - host: host1.example.com
    key_file: /path/to/key
    timeout: 5
    batch_size: 10000
    features: # enable/disable per host
      routes: false
  - host: host2.example.com:2233
    username: exporter
    password: secret

features:
  system: true
  environment: true
  interfaces: true
  optics: true
  routes: true
  wifi: true
```

# Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)
* Logrus Logging library (https://github.com/sirupsen/logrus)

# License
(c) slashdoom (Patrick Ryon), 2022. Licensed under [MIT](LICENSE) license.

# Prometheus
see https://prometheus.io/