package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudruntime"
)

// Runtime returns the runtime config for the current context.
func Runtime(ctx context.Context) cloudruntime.Config {
	config, ok := cloudruntime.GetConfig(ctx)
	if !ok {
		panic("cloudrunner.Runtime must be called with a context from cloudrunner.Run")
	}
	return config
}
