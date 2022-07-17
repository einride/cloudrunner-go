package cloudconfig

// Option is a configuration option.
type Option func(*Config)

// WithAdditionalSpec includes an additional specification in the config loading.
func WithAdditionalSpec(name string, spec interface{}) Option {
	return func(config *Config) {
		config.configSpecs = append(config.configSpecs, &configSpec{
			name: name,
			spec: spec,
		})
	}
}

// WithEnvPrefix sets the environment prefix to use for config loading.
func WithEnvPrefix(envPrefix string) Option {
	return func(config *Config) {
		config.envPrefix = envPrefix
	}
}

// WithYAMLServiceSpecificationFile sets the YAML service specification file to load environment variables from.
func WithYAMLServiceSpecificationFile(filename string) Option {
	return func(config *Config) {
		config.yamlServiceSpecificationFilename = filename
	}
}

// WithOptionalSecrets overrides all secrets to be optional.
func WithOptionalSecrets() Option {
	return func(config *Config) {
		config.optionalSecrets = true
	}
}
