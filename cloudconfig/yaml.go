package cloudconfig

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type env struct {
	Name      string
	Value     string
	ValueFrom struct {
		SecretKeyRef struct {
			Key  string
			Name string
		} `yaml:"secretKeyRef"`
	} `yaml:"valueFrom"`
}

// readYAMLServiceSpecificationFile reads the service name and the env of the first container
// from a YAML service, worker pool or job specification file.
func readYAMLServiceSpecificationFile(name string) (serviceName string, envs []env, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("set env from YAML service/job specification file %s: %w", name, err)
		}
	}()
	var kind struct {
		Kind string
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return "", nil, err
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&kind); err != nil {
		return "", nil, err
	}
	switch kind.Kind {
	case "Service", "WorkerPool": // Cloud Run Services and Worker Pools
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
			return "", nil, err
		}
		containers := config.Spec.Template.Spec.Containers
		if len(containers) == 0 || len(containers) > 10 {
			return "", nil, fmt.Errorf("unexpected number of containers: %d", len(containers))
		}
		return config.Metadata.Name, containers[0].Env, nil
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
			return "", nil, err
		}
		containers := config.Spec.Template.Spec.Template.Spec.Containers
		if len(containers) == 0 || len(containers) > 10 {
			return "", nil, fmt.Errorf("unexpected number of containers: %d", len(containers))
		}
		return config.Metadata.Name, containers[0].Env, nil
	default:
		return "", nil, fmt.Errorf("unknown config kind: %s", kind.Kind)
	}
}
