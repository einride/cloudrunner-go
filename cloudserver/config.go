package cloudserver

import (
	"time"
)

// Config provides config for gRPC and HTTP servers.
type Config struct {
	// Timeout of all requests to the servers.
	// Defaults to 10 seconds below the default Cloud Run timeout for managed services.
	Timeout time.Duration `default:"290s"`
}
