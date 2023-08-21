package cloudprofiler

import (
	"errors"
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

	var cloudConfig cloudruntime.Config
	if err := cloudConfig.Autodetect(); err != nil {
		return fmt.Errorf("start profiler: %w", err)
	}

	var svcConfig profiler.Config
	switch {
	case cloudConfig.Service != "":
		var err error
		svcConfig, err = cloudRunServiceConfig(cloudConfig)
		if err != nil {
			return fmt.Errorf("start profiler: %w", err)
		}
	case cloudConfig.Job != "":
		var err error
		svcConfig, err = cloudRunJobConfig(cloudConfig)
		if err != nil {
			return fmt.Errorf("start profiler: %w", err)
		}
	default:
		return errors.New("unable to autodetect runtime environment")
	}

	svcConfig.MutexProfiling = config.MutexProfiling
	svcConfig.AllocForceGC = config.AllocForceGC

	if err := profilerStart(svcConfig); err != nil {
		return fmt.Errorf("start profiler: %w", err)
	}
	return nil
}

func cloudRunServiceConfig(cloudCfg cloudruntime.Config) (profiler.Config, error) {
	return profiler.Config{
		ProjectID:      cloudCfg.ProjectID,
		Service:        cloudCfg.Service,
		ServiceVersion: cloudCfg.ServiceVersion,
	}, nil
}

func cloudRunJobConfig(cloudCfg cloudruntime.Config) (profiler.Config, error) {
	return profiler.Config{
		ProjectID:      cloudCfg.ProjectID,
		Service:        cloudCfg.Job,
		ServiceVersion: cloudCfg.Execution,
	}, nil
}
