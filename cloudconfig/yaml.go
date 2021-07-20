package cloudconfig

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func setEnvFromYAMLServiceSpecificationFile(name string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("set env from YAML service specification file %s: %w", name, err)
		}
	}()
	var config struct {
		Metadata struct {
			Name string
		}
		Spec struct {
			Template struct {
				Spec struct {
					Containers []struct {
						Env []struct {
							Name  string
							Value string
						}
					}
				}
			}
		}
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
		return err
	}
	if len(config.Spec.Template.Spec.Containers) != 1 {
		return fmt.Errorf("unexpected number of containers: %d", len(config.Spec.Template.Spec.Containers))
	}
	if config.Metadata.Name != "" {
		if err := os.Setenv("K_SERVICE", config.Metadata.Name); err != nil {
			return err
		}
	}
	for _, env := range config.Spec.Template.Spec.Containers[0].Env {
		// Prefer variables from local environment.
		if _, ok := os.LookupEnv(env.Name); !ok {
			if err := os.Setenv(env.Name, env.Value); err != nil {
				return err
			}
		}
	}
	return nil
}
