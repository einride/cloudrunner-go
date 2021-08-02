package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudruntime"
)

// Service returns the service config for the current context.
func Service(ctx context.Context) ServiceConfig {
	runCtx, ok := getRunContext(ctx)
	if !ok {
		panic("cloudrunner.Logger must be called with a context from cloudrunner.Run")
	}
	return runCtx.runConfig.Service
}

// ServiceConfig contains generic service configuration.
type ServiceConfig struct {
	// Port is the port the service is listening on.
	Port int `env:"PORT" default:"8080"`
	// Name is the name of the service.
	Name string `env:"K_SERVICE"`
	// Revision of the service, as assigned by a Knative runtime.
	Revision string `env:"K_REVISION"`
	// Configuration of the service, as assigned by a Knative runtime.
	Configuration string `env:"K_CONFIGURATION"`
	// ProjectID is the GCP project ID the service is running in.
	// In production, defaults to the project where the service is deployed.
	ProjectID string `env:"GOOGLE_CLOUD_PROJECT"`
	// Account is the service account used by the service.
	// In production, defaults to the default service account of the running service.
	Account string
	// Version is the version of the service.
	// Defaults to ServiceVersion (which can be set during build-time).
	Version string
}

func (c *ServiceConfig) loadFromRuntime() {
	if projectID, ok := cloudruntime.ProjectID(); ok {
		c.ProjectID = projectID
	}
	if serviceVersion, ok := cloudruntime.ServiceVersion(); ok {
		c.Version = serviceVersion
	}
	if serviceAccount, ok := cloudruntime.ServiceAccount(); ok {
		c.Account = serviceAccount
	}
	if service, ok := cloudruntime.Service(); ok {
		c.Name = service
	}
	if revision, ok := cloudruntime.Revision(); ok {
		c.Revision = revision
	}
	if configuration, ok := cloudruntime.Configuration(); ok {
		c.Configuration = configuration
	}
}
