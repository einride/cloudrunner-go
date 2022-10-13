package cloudprofiler

import (
	"fmt"

	"cloud.google.com/go/profiler"
	"go.einride.tech/cloudrunner/cloudruntime"
)

// shims for unit testing.
//
//nolint:gochecknoglobals
var (
	profilerStart = profiler.Start
)

// Config configures the use of Google Cloud Profiler.
type Config struct {
	// Enabled indicates if the profiler should be enabled.
	Enabled bool `onGCE:"true"`
	// MutexProfiling indicates if mutex profiling should be enabled.
	MutexProfiling bool
	// AllocForceGC indicates if GC should be forced before allocation profiling snapshots are taken.
	AllocForceGC bool `default:"true"`
}

// Start the profiler according to the provided Config.
func Start(config Config) error {
	if !config.Enabled {
		return nil
	}
	service, ok := cloudruntime.Service()
	if !ok {
		return fmt.Errorf("start profiler: missing service")
	}
	projectID, ok := cloudruntime.ProjectID()
	if !ok {
		return fmt.Errorf("start profiler: missing project ID")
	}
	serviceVersion, ok := cloudruntime.ServiceVersion()
	if !ok {
		return fmt.Errorf("start profiler: missing service version")
	}
	if err := profilerStart(profiler.Config{
		ProjectID:      projectID,
		Service:        service,
		ServiceVersion: serviceVersion,
		MutexProfiling: config.MutexProfiling,
		AllocForceGC:   config.AllocForceGC,
	}); err != nil {
		return fmt.Errorf("start profiler: %w", err)
	}
	return nil
}
