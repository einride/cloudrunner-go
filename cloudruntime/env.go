package cloudruntime

import (
	"os"
	"path"
	"strconv"
)

// Service returns the service name of the current runtime.
func Service() (string, bool) {
	if kService, ok := os.LookupEnv("K_SERVICE"); ok {
		return kService, true
	}
	// Default to the name of the entrypoint command.
	return path.Base(os.Args[0]), true
}

// Revision returns the service revision of the current runtime.
func Revision() (string, bool) {
	return os.LookupEnv("K_REVISION")
}

// Configuration returns the service configuration of the current runtime.
func Configuration() (string, bool) {
	return os.LookupEnv("K_CONFIGURATION")
}

// Port returns the service port of the current runtime.
func Port() (int, bool) {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	return port, err == nil
}
