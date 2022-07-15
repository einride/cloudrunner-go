// Package cloudyaml contains primitives for working with YAML service specifications.
package cloudyaml

import (
	"bytes"
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"gopkg.in/yaml.v3"
)

// ResolveEnvFromFile resolves the environment specified in the provided YAML service specification file.
// Secrets are resolved in the provided Google Cloud project.
func ResolveEnvFromFile(ctx context.Context, project, filename string) (_ []string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("resolve env from YAML service specification file %s: %w", filename, err)
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
							Name      string
							Value     string
							ValueFrom struct {
								SecretKeyRef struct {
									Name string
									Key  string
								}
							}
						}
					}
				}
			}
		}
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
		return nil, err
	}
	if len(config.Spec.Template.Spec.Containers) != 1 {
		return nil, fmt.Errorf("unexpected number of containers: %d", len(config.Spec.Template.Spec.Containers))
	}
	result := make([]string, 0, 100)
	if config.Metadata.Name != "" {
		result = append(result, "K_SERVICE="+config.Metadata.Name)
	}
	var secretClient *secretmanager.Client
	for _, env := range config.Spec.Template.Spec.Containers[0].Env {
		if env.Value != "" {
			result = append(result, env.Name+"="+env.Value)
			continue
		}
		if env.ValueFrom.SecretKeyRef.Name != "" && env.ValueFrom.SecretKeyRef.Key != "" {
			if secretClient == nil {
				secretClient, err = secretmanager.NewClient(ctx)
				if err != nil {
					return nil, err
				}
			}
			response, err := secretClient.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
				Name: fmt.Sprintf(
					"projects/%s/secrets/%s/versions/%s",
					project,
					env.ValueFrom.SecretKeyRef.Name,
					env.ValueFrom.SecretKeyRef.Key,
				),
			})
			if err != nil {
				return nil, err
			}
			result = append(result, env.Name+"="+string(response.GetPayload().GetData()))
		}
	}
	return result, nil
}
