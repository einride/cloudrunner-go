package cloudruntime

import (
	"context"
	"encoding/json"
	"os"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2/google"
)

// shims for unit testing.
var (
	metadataOnGCE                = metadata.OnGCE
	metadataProjectID            = metadata.ProjectID
	metadataEmail                = metadata.Email
	googleFindDefaultCredentials = google.FindDefaultCredentials
)

// ProjectID returns the Google Cloud Project ID of the current runtime.
func ProjectID() (string, bool) {
	if !metadataOnGCE() {
		if projectFromEnv, ok := os.LookupEnv("GOOGLE_CLOUD_PROJECT"); ok {
			return projectFromEnv, true
		}
		return projectIDFromDefaultCredentials()
	}
	projectID, err := metadataProjectID()
	return projectID, err == nil
}

// ServiceAccount returns the default service account of the current runtime.
func ServiceAccount() (string, bool) {
	if !metadataOnGCE() {
		return serviceAccountFromDefaultCredentials()
	}
	serviceAccount, err := metadataEmail("default")
	return serviceAccount, err == nil
}

func projectIDFromDefaultCredentials() (string, bool) {
	defaultCredentials, err := googleFindDefaultCredentials(context.Background())
	if err != nil {
		return "", false
	}
	return defaultCredentials.ProjectID, defaultCredentials.ProjectID != ""
}

func serviceAccountFromDefaultCredentials() (string, bool) {
	defaultCredentials, err := googleFindDefaultCredentials(context.Background())
	if err != nil || defaultCredentials.JSON == nil {
		return "", false
	}
	var credentials struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
	}
	if err := json.Unmarshal(defaultCredentials.JSON, &credentials); err != nil {
		return "", false
	}
	if credentials.Type != "service_account" {
		return "", false
	}
	return credentials.ClientEmail, credentials.ClientEmail != ""
}
