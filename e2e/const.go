package e2e

import (
	"time"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 300
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 60
	operatorName         = "kubevirt-image-service"
)
