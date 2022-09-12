package cloudruntime

import (
	"os"
	"runtime/debug"
)

//nolint: gochecknoglobals
var serviceVersion string

// ServiceVersionFromLinkerFlags returns the exact value of the variable:
//
//  go.einride.tech/cloudrunner/cloudruntime.serviceVersion
//
// This variable can be set during build-time to provide a default value for the service version.
//
// Example:
//
//  go build -ldflags="-X 'go.einride.tech/cloudrunner/cloudruntime.serviceVersion=v1.0.0'"
func ServiceVersionFromLinkerFlags() string {
	return serviceVersion
}

// ServiceVersionFromBuildInfo returns the VCS revision from the embedded build info.
func ServiceVersionFromBuildInfo() (string, bool) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value, true
		}
	}
	return "", false
}

// ServiceVersion returns the service version of the current runtime.
// The service version is taken from, in order of precedence:
// - the "SERVICE_VERSION" environment variable
// - the go.einride.tech/cloudrunner/cloudruntime.serviceVersion variable (must be set at build-time)
// - the "K_REVISION" environment variable
// - no version.
func ServiceVersion() (string, bool) {
	// TODO: Remove support for everything except version from build info.
	if serviceVersionFromEnv, ok := os.LookupEnv("SERVICE_VERSION"); ok {
		return serviceVersionFromEnv, ok
	}
	if ServiceVersionFromLinkerFlags() != "" {
		return ServiceVersionFromLinkerFlags(), true
	}
	if version, ok := ServiceVersionFromBuildInfo(); ok {
		return version, true
	}
	if revision, ok := os.LookupEnv("K_REVISION"); ok {
		return revision, true
	}
	return "", false
}
