package confluent

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	registry "github.com/faustbrian/go-schema-registry/v3"
)

func TestPrivacyCredentialFailure(t *testing.T) {
	cause := errors.New("application-private-detail")
	config := internalConfig(roundTripperFunction(func(*http.Request) (*http.Response, error) {
		t.Fatal("transport called after credential failure")
		return nil, nil
	}))
	config.Credentials = credentialFunction(func(context.Context) (string, error) { return "", cause })
	provider, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Resolve(context.Background(), registry.ByProviderID(registry.ProviderID{Provider: ProviderName, Scope: config.Scope, Value: "1"}))
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Errorf("private credential diagnostic: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("credential cause lost")
	}
}

func TestPrivacyUnknownSchemaType(t *testing.T) {
	config := internalConfig(roundTripperFunction(func(*http.Request) (*http.Response, error) { t.Fatal("unexpected transport"); return nil, nil }))
	config.Canonicalizers = map[registry.Format]registry.Canonicalizer{registry.Format("application-private-detail"): canonicalizerFunction(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil })}
	_, err := New(config)
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatalf("private schema-type diagnostic: %v", err)
	}
}
