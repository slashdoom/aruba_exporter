package wireless

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/slashdoom/aruba_exporter/rpc"
	"github.com/slashdoom/aruba_exporter/util"
	
	log "github.com/sirupsen/logrus"
)