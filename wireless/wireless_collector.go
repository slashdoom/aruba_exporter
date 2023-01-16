package wireless

import (
	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix string = "aruba_system_"