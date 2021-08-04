package cloudruntime

import "context"

// WithConfig adds the provided runtime Config to the current context.
func WithConfig(ctx context.Context, config Config) context.Context {
	return context.WithValue(ctx, configContextKey{}, config)
}

// GetConfig gets the runtime Config from the current context.
func GetConfig(ctx context.Context) (Config, bool) {
	result, ok := ctx.Value(configContextKey{}).(Config)
	return result, ok
}

type configContextKey struct{}

// Config is the runtime config for the service.
type Config struct {
	// Port is the port the service is listening on.
	Port int `env:"PORT" default:"8080"`
	// Service is the name of the service.
	Service string `env:"K_SERVICE"`
	// Revision of the service, as assigned by a Knative runtime.
	Revision string `env:"K_REVISION"`
	// Configuration of the service, as assigned by a Knative runtime.
	Configuration string `env:"K_CONFIGURATION"`
	// ProjectID is the GCP project ID the service is running in.
	// In production, defaults to the project where the service is deployed.
	ProjectID string `env:"GOOGLE_CLOUD_PROJECT"`
	// ServiceAccount is the service account used by the service.
	// In production, defaults to the default service account of the running service.
	ServiceAccount string
	// ServiceVersion is the version of the service.
	ServiceVersion string `env:"SERVICE_VERSION"`
}

// Autodetect the runtime config.
func (c *Config) Autodetect() error {
	if projectID, ok := ProjectID(); ok {
		c.ProjectID = projectID
	}
	if serviceVersion, ok := ServiceVersion(); ok {
		c.ServiceVersion = serviceVersion
	}
	if serviceAccount, ok := ServiceAccount(); ok {
		c.ServiceAccount = serviceAccount
	}
	if service, ok := Service(); ok {
		c.Service = service
	}
	if revision, ok := Revision(); ok {
		c.Revision = revision
	}
	if configuration, ok := Configuration(); ok {
		c.Configuration = configuration
	}
	return nil
}
