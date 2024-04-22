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
			err = fmt.Errorf("set env from YAML service/job specification file %s: %w", name, err)
		}
	}()
	var kind struct {
		Kind string
	}
	type env struct {
		Name  string
		Value string
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&kind); err != nil {
		return err
	}
	var envs []env
	switch kind.Kind {
	case "Service": // Cloud Run Services
		var config struct {
			Metadata struct {
				Name string
			}
			Spec struct {
				Template struct {
					Spec struct {
						Containers []struct {
							Env []env
						}
					}
				}
			}
		}
		if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
			return err
		}
		containers := config.Spec.Template.Spec.Containers
		if len(containers) == 0 || len(containers) > 10 {
			return fmt.Errorf("unexpected number of containers: %d", len(containers))
		}
		if config.Metadata.Name != "" {
			if err := os.Setenv("K_SERVICE", config.Metadata.Name); err != nil {
				return err
			}
		}
		envs = containers[0].Env
	case "Job": // Cloud Run Jobs
		var config struct {
			Metadata struct {
				Name string
			}
			Spec struct {
				Template struct {
					Spec struct {
						Template struct {
							Spec struct {
								Containers []struct {
									Env []env
								}
							}
						}
					}
				}
			}
		}
		if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
			return err
		}
		containers := config.Spec.Template.Spec.Template.Spec.Containers
		if len(containers) == 0 || len(containers) > 10 {
			return fmt.Errorf("unexpected number of containers: %d", len(containers))
		}
		if config.Metadata.Name != "" {
			if err := os.Setenv("K_SERVICE", config.Metadata.Name); err != nil {
				return err
			}
		}
		envs = containers[0].Env
	default:
		return fmt.Errorf("unknown config kind: %s", kind.Kind)
	}
	for _, env := range envs {
		// Prefer variables from local environment.
		if _, ok := os.LookupEnv(env.Name); !ok {
			if err := os.Setenv(env.Name, env.Value); err != nil {
				return err
			}
		}
	}
	return nil
}
