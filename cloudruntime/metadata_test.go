package cloudruntime

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/oauth2/google"
	"gotest.tools/v3/assert"
)

func TestProjectID(t *testing.T) {
	t.Run("on GCE", func(t *testing.T) {
		const expected = "foo"
		withOnGCE(t, true)
		withMetadataProjectID(t, expected, nil)
		actual, ok := ProjectID()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("default credentials", func(t *testing.T) {
		const expected = "foo"
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, &google.Credentials{ProjectID: expected}, nil)
		actual, ok := ProjectID()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("missing default credentials", func(t *testing.T) {
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, nil, fmt.Errorf("not found"))
		actual, ok := ProjectID()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})
}

func TestServiceAccount(t *testing.T) {
	t.Run("on GCE", func(t *testing.T) {
		const expected = "foo@bar.iam.gserviceaccount.com"
		withOnGCE(t, true)
		withMetadataEmail(t, "default", expected, nil)
		actual, ok := ServiceAccount()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("default credentials", func(t *testing.T) {
		const expected = "foo@bar.iam.gserviceaccount.com"
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, &google.Credentials{
			JSON: []byte(fmt.Sprintf(`{"type":"service_account","client_email":"%s"}`, expected)),
		}, nil)
		actual, ok := ServiceAccount()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("user default credentials", func(t *testing.T) {
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, &google.Credentials{
			JSON: []byte(`{"type":"user","client_email":"foo@example.com"}`),
		}, nil)
		actual, ok := ServiceAccount()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})

	t.Run("missing default credentials", func(t *testing.T) {
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, nil, fmt.Errorf("not found"))
		actual, ok := ServiceAccount()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})

	t.Run("invalid default credentials", func(t *testing.T) {
		withOnGCE(t, false)
		withGoogleDefaultCredentials(t, &google.Credentials{JSON: []byte(`foo`)}, nil)
		actual, ok := ServiceAccount()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})
}

func withGoogleDefaultCredentials(t *testing.T, credentials *google.Credentials, err error) {
	prev := googleFindDefaultCredentials
	googleFindDefaultCredentials = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
		return credentials, err
	}
	t.Cleanup(func() {
		googleFindDefaultCredentials = prev
	})
}

func withMetadataProjectID(t *testing.T, value string, err error) {
	prev := metadataProjectID
	metadataProjectID = func() (string, error) {
		return value, err
	}
	t.Cleanup(func() {
		metadataProjectID = prev
	})
}

func withMetadataEmail(t *testing.T, expectedServiceAccount, value string, err error) {
	prev := metadataEmail
	metadataEmail = func(serviceAccount string) (string, error) {
		assert.Equal(t, expectedServiceAccount, serviceAccount)
		return value, err
	}
	t.Cleanup(func() {
		metadataEmail = prev
	})
}

func withOnGCE(t *testing.T, value bool) {
	prev := metadataOnGCE
	metadataOnGCE = func() bool {
		return value
	}
	t.Cleanup(func() {
		metadataOnGCE = prev
	})
}
