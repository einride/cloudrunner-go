package cloudrunner

import (
	"fmt"

	"go.einride.tech/cloudrunner/cloudconfig"
)

// ValidateConfig validates a YAML service specification file against the config
// that options register, like Run does when started with -config and -validate.
//
// Unlike Run, it reads values only from the file and never from the process environment,
// which it leaves untouched. Secrets are optional, since their values are not in the file.
func ValidateConfig(yamlServiceSpecificationFile string, options ...Option) error {
	var run runContext
	for _, option := range options {
		option(&run)
	}
	run.configOptions = append(
		run.configOptions,
		cloudconfig.WithYAMLServiceSpecificationFile(yamlServiceSpecificationFile),
		cloudconfig.WithOptionalSecrets(),
		// A lookup that finds nothing, so the config is validated against the file alone: the
		// file's env is kept in memory instead of being set in the process environment, and
		// variables already in the process environment can't fill in values the file is missing.
		cloudconfig.WithLookupEnv(func(string) (string, bool) { return "", false }),
	)
	config, err := cloudconfig.New("cloudrunner", &run.config, run.configOptions...)
	if err != nil {
		return fmt.Errorf("cloudrunner.ValidateConfig: %w", err)
	}
	if err := config.Load(); err != nil {
		return fmt.Errorf("cloudrunner.ValidateConfig: %w", err)
	}
	return nil
}
