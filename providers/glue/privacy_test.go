package glue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	awsglue "github.com/aws/aws-sdk-go-v2/service/glue"
	registry "github.com/faustbrian/go-schema-registry/v2"
)

func TestPrivacySDKDiagnostic(t *testing.T) {
	cause := errors.New("application-private-detail")
	provider := internalProvider(t, &apiFunction{version: func(context.Context, *awsglue.GetSchemaVersionInput) (*awsglue.GetSchemaVersionOutput, error) {
		return nil, cause
	}})
	_, err := provider.Resolve(context.Background(), registry.ByProviderID(registry.ProviderID{Provider: ProviderName, Scope: "scope", Value: internalVersionID}))
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Errorf("private SDK diagnostic: %v", err)
	}
	if !errors.Is(err, cause) || !errors.Is(err, registry.ErrUnavailable) {
		t.Fatal("SDK classification/cause lost")
	}
}

func TestPrivacySDKCancellationDiagnostic(t *testing.T) {
	cause := errors.Join(context.Canceled, errors.New("application-private-detail"))
	provider := internalProvider(t, &apiFunction{version: func(context.Context, *awsglue.GetSchemaVersionInput) (*awsglue.GetSchemaVersionOutput, error) {
		return nil, cause
	}})
	_, err := provider.Resolve(context.Background(), registry.ByProviderID(registry.ProviderID{Provider: ProviderName, Scope: "scope", Value: internalVersionID}))
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Errorf("private cancellation diagnostic: %v", err)
	}
	if !errors.Is(err, cause) || !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation/cause lost")
	}
	children, ok := err.(interface{ Unwrap() []error })
	want := cause.(interface{ Unwrap() []error }).Unwrap()
	if !ok || len(children.Unwrap()) != len(want) || children.Unwrap()[0] != want[0] || children.Unwrap()[1] != want[1] {
		t.Fatal("joined cancellation topology changed")
	}
}

func TestPrivacyWrappedSDKCancellationTopology(t *testing.T) {
	for _, category := range []error{context.Canceled, context.DeadlineExceeded} {
		cause := fmt.Errorf("application-private-detail: %w", category)
		err := classifyError(cause)
		if errors.Unwrap(err) != category || !errors.Is(err, cause) || strings.Contains(err.Error(), "application-private-detail") {
			t.Fatal("wrapped cancellation topology/privacy changed")
		}
		if classifyError(category) != category {
			t.Fatal("bare cancellation identity changed")
		}
	}
}

type privacyContextCause struct{ category error }

func (err *privacyContextCause) Error() string        { return "application-private-detail" }
func (err *privacyContextCause) Is(target error) bool { return target == err.category }

func TestPrivacyTypedSDKCancellationIdentity(t *testing.T) {
	cause := &privacyContextCause{category: context.Canceled}
	err := classifyError(cause)
	var original *privacyContextCause
	if !errors.As(err, &original) || original != cause || !errors.Is(err, cause) || errors.Unwrap(err) != nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatal("typed context privacy/inspection/topology lost")
	}
}

func TestPrivacyUnknownSchemaType(t *testing.T) {
	_, err := New(Config{API: noCallAPI(t), Scope: "scope", RequestTimeout: 1, MaxConcurrent: 1, Canonicalizers: map[registry.Format]registry.Canonicalizer{registry.Format("application-private-detail"): canonicalizerFunction(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil })}})
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatalf("private schema-type diagnostic: %v", err)
	}
}
