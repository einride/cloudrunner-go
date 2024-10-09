package cloudconfig

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"go.opentelemetry.io/otel/codes"
)

// envPrefix can be set during build-time to append a prefix to all environment variables loaded into the RunConfig.
//
//nolint:gochecknoglobals
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
	optionalSecrets                  bool
}

type configSpec struct {
	name       string
	spec       interface{}
	fieldSpecs []fieldSpec
}

// Load values into the config.
func (c *Config) Load() error {
	if c.yamlServiceSpecificationFilename != "" {
		envs, err := getEnvFromYAMLServiceSpecificationFile(c.yamlServiceSpecificationFilename)
		if err != nil {
			return err
		}
		if err := validateEnvSecretTags(envs, c.configSpecs); err != nil {
			return err
		}
		for _, e := range envs {
			if err := os.Setenv(e.Name, e.Value); err != nil {
				return err
			}
		}
	}
	for _, cs := range c.configSpecs {
		if err := c.process(cs.fieldSpecs); err != nil {
			return err
		}
	}
	return nil
}

func validateEnvSecretTags(envs []env, configSpecs []*configSpec) error {
	for _, env := range envs {
		if env.ValueFrom.SecretKeyRef.Key == "" && env.ValueFrom.SecretKeyRef.Name == "" {
			continue
		}
		for _, spec := range configSpecs {
			for _, f := range spec.fieldSpecs {
				if f.Key == env.Name {
					if !f.Secret {
						return fmt.Errorf("field %s does not have the correct secret tag", f.Name)
					}
				}
			}
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

// LogValue implements [slog.LogValuer].
func (c *Config) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(c.configSpecs))
	for _, configSpec := range c.configSpecs {
		attrs = append(attrs, slog.Any(configSpec.name, fieldSpecsValue(configSpec.fieldSpecs)))
	}
	return slog.GroupValue(attrs...)
}

type fieldSpecsValue []fieldSpec

func (fsv fieldSpecsValue) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(fsv))
	for _, fs := range fsv {
		if fs.Secret {
			attrs = append(attrs, slog.String(fs.Key, "<secret>"))
			continue
		}
		switch value := fs.Value.Interface().(type) {
		case time.Duration:
			attrs = append(attrs, slog.Duration(fs.Key, value))
		case []codes.Code:
			logValue := make([]string, 0, len(value))
			for _, code := range value {
				logValue = append(logValue, code.String())
			}
			attrs = append(attrs, slog.Any(fs.Key, logValue))
		case map[codes.Code]slog.Level:
			logValue := make(map[string]string, len(value))
			for code, level := range value {
				logValue[code.String()] = level.String()
			}
			attrs = append(attrs, slog.Any(fs.Key, logValue))
		default:
			attrs = append(attrs, slog.Any(fs.Key, fs.Value.Interface()))
		}
	}
	return slog.GroupValue(attrs...)
}
