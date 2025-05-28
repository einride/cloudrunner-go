package cloudruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2/google"
)

// shims for unit testing.
//
//nolint:gochecknoglobals
var (
	metadataOnGCE                = metadata.OnGCE
	metadataProjectIDWithContext = metadata.ProjectIDWithContext
	metadataEmailWithContext     = metadata.EmailWithContext
	googleFindDefaultCredentials = google.FindDefaultCredentials
)

// ProjectID returns the Google Cloud Project ID of the current runtime.
// Deprecated: Use the context-based [ResolveProjectID] function.
func ProjectID() (string, bool) {
	return ResolveProjectID(context.Background())
}

// ResolveProjectID resolves the Google Cloud Project ID of the current runtime.
func ResolveProjectID(ctx context.Context) (string, bool) {
	if !metadataOnGCE() {
		if projectFromEnv, ok := os.LookupEnv("GOOGLE_CLOUD_PROJECT"); ok {
			return projectFromEnv, true
		}
		return projectIDFromDefaultCredentials(ctx)
	}
	projectID, err := metadataProjectIDWithContext(ctx)
	return projectID, err == nil
}

// ServiceAccount returns the default service account of the current runtime.
// Deprecated: Use the context-based [ResolveServiceAccount] function.
func ServiceAccount() (string, bool) {
	return ResolveServiceAccount(context.Background())
}

// ResolveServiceAccount resolves the default service account of the current runtime.
func ResolveServiceAccount(ctx context.Context) (string, bool) {
	if !metadataOnGCE() {
		return serviceAccountFromDefaultCredentials(ctx)
	}
	serviceAccount, err := metadataEmailWithContext(ctx, "default")
	return serviceAccount, err == nil
}

func projectIDFromDefaultCredentials(ctx context.Context) (string, bool) {
	defaultCredentials, err := googleFindDefaultCredentials(ctx)
	if err != nil {
		return "", false
	}
	return defaultCredentials.ProjectID, defaultCredentials.ProjectID != ""
}

func serviceAccountFromDefaultCredentials(ctx context.Context) (string, bool) {
	defaultCredentials, err := googleFindDefaultCredentials(ctx)
	if err != nil || defaultCredentials.JSON == nil {
		return "", false
	}
	var credentials struct {
		Type                           string `json:"type"`
		ClientEmail                    string `json:"client_email"`
		ServiceAccountImpersonationURL string `json:"service_account_impersonation_url"`
	}
	if err := json.Unmarshal(defaultCredentials.JSON, &credentials); err != nil {
		return "", false
	}
	switch credentials.Type {
	case "service_account":
		return credentials.ClientEmail, credentials.ClientEmail != ""
	case "impersonated_service_account":
		sa, err := extractServiceAccountFromImpersonationURL(credentials.ServiceAccountImpersonationURL)
		if err != nil {
			return "", false
		}
		return sa, true
	}
	return "", false
}

// extractServiceAccountFromImpersonationURL parses a URL like
// ".../serviceAccounts/sa@project.iam.gserviceaccount.com:generateAccessToken"
// and returns "sa@project.iam.gserviceaccount.com".
func extractServiceAccountFromImpersonationURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "serviceAccounts" && i+1 < len(parts) {
			sa := parts[i+1]
			if idx := strings.Index(sa, ":"); idx != -1 {
				sa = sa[:idx]
			}
			return sa, nil
		}
	}
	return "", fmt.Errorf("service account not found in URL")
}
