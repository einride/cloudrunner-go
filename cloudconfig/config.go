package cloudconfig

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// envPrefix can be set during build-time to append a prefix to all environment variables loaded into the RunConfig.
var envPrefix string

// New creates a new Config with the provided name, specification and options.
func New(name string, spec interface{}, options ...Option) (*Config, error) {
	config := Config{
		configSpecs: []*configSpec{
			{name: name, spec: spec},
		},
		envPrefix: envPrefix,
	}
	for _, option := range options {
		option(&config)
	}
	for _, configSpec := range config.configSpecs {
		fieldSpecs, err := collectFieldSpecs(config.envPrefix, configSpec.spec)
		if err != nil {
			return nil, err
		}
		configSpec.fieldSpecs = fieldSpecs
	}
	return &config, nil
}

// Config is a config.
type Config struct {
	configSpecs                      []*configSpec
	envPrefix                        string
	yamlServiceSpecificationFilename string
}

type configSpec struct {
	name       string
	spec       interface{}
	fieldSpecs []fieldSpec
}

// Load values into the config.
func (c *Config) Load() error {
	if c.yamlServiceSpecificationFilename != "" {
		if err := setEnvFromYAMLServiceSpecificationFile(c.yamlServiceSpecificationFilename); err != nil {
			return err
		}
	}
	for _, cs := range c.configSpecs {
		if err := process(cs.fieldSpecs); err != nil {
			return err
		}
	}
	return nil
}

// PrintUsage prints usage of the config to the provided io.Writer.
func (c *Config) PrintUsage(w io.Writer) {
	tabs := tabwriter.NewWriter(w, 1, 0, 4, ' ', 0)
	_, _ = fmt.Fprintf(tabs, "CONFIG\tENV\tTYPE\tDEFAULT\tON GCE\n")
	for _, cs := range c.configSpecs {
		for _, fs := range cs.fieldSpecs {
			_, _ = fmt.Fprintf(
				tabs,
				"%v\t%v\t%v\t%v\t%v\n",
				cs.name,
				fs.Key,
				fs.Value.Type(),
				fs.Tags.Get("default"),
				fs.Tags.Get("onGCE"),
			)
		}
	}
	_ = tabs.Flush()
}
