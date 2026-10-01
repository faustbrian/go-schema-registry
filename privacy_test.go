package schemaregistry_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	registry "github.com/faustbrian/go-schema-registry"
)

const privacyMarker = "application-private-detail"

type privacyCause struct{}

func (*privacyCause) Error() string { return privacyMarker }

func assertPrivateDiagnostic(t *testing.T, err, cause, category error) {
	t.Helper()
	if err == nil || strings.Contains(fmt.Sprint(err), privacyMarker) || strings.Contains(fmt.Sprintf("%+v", err), privacyMarker) {
		t.Errorf("default error exposes private detail: %v", err)
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Errorf("original cause identity lost: %v", err)
	}
	if category != nil && !errors.Is(err, category) {
		t.Errorf("category lost: %v", err)
	}
	if original, ok := cause.(*privacyCause); ok {
		var inspected *privacyCause
		if !errors.As(err, &inspected) || inspected != original {
			t.Fatal("typed collaborator identity lost")
		}
	}
}

func TestPrivacyCompileAndReferences(t *testing.T) {
	cause := &privacyCause{}
	_, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"string"`)}, canonicalizerFunc(func(context.Context, registry.Definition) ([]byte, error) { return nil, cause }))
	assertPrivateDiagnostic(t, err, cause, registry.ErrInvalidSchema)
	children, ok := err.(interface{ Unwrap() []error })
	if !ok || len(children.Unwrap()) != 2 || children.Unwrap()[0] != registry.ErrInvalidSchema || children.Unwrap()[1] != cause {
		t.Fatal("canonicalization lost direct multi-cause topology")
	}
	var typed *privacyCause
	if !errors.As(err, &typed) || typed != cause {
		t.Fatal("typed cause unavailable")
	}
	leaf, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"string"`)}, canonicalizerFunc(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil }))
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"string"`), References: []registry.Reference{{Name: privacyMarker, Fingerprint: leaf.Fingerprint()}, {Name: privacyMarker, Fingerprint: leaf.Fingerprint()}}}, canonicalizerFunc(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil }))
	assertPrivateDiagnostic(t, err, nil, registry.ErrInvalidSchema)
	coordinate := registry.ReferenceCoordinate{Subject: registry.Subject{Name: privacyMarker}, Version: registry.Version{Number: 1}}
	_, err = registry.BuildReferenceGraph(context.Background(), []registry.ReferenceCoordinate{coordinate}, referenceResolverFunc(func(context.Context, registry.ReferenceCoordinate) (registry.ReferenceDocument, error) {
		return registry.ReferenceDocument{}, cause
	}), registry.GraphLimits{MaxSchemas: 2, MaxDepth: 2, MaxReferences: 2})
	assertPrivateDiagnostic(t, err, cause, nil)
	if errors.Unwrap(err) != cause {
		t.Fatal("reference resolver lost direct cause")
	}
	_, err = registry.BuildReferenceGraph(context.Background(), []registry.ReferenceCoordinate{coordinate}, referenceResolverFunc(func(context.Context, registry.ReferenceCoordinate) (registry.ReferenceDocument, error) {
		return registry.ReferenceDocument{}, registry.ErrNotFound
	}), registry.GraphLimits{MaxSchemas: 2, MaxDepth: 2, MaxReferences: 2})
	missing, ok := err.(interface{ Unwrap() []error })
	if !ok || len(missing.Unwrap()) != 2 || missing.Unwrap()[0] != registry.ErrReferenceMissing || missing.Unwrap()[1] != registry.ErrNotFound {
		t.Fatal("graph missing reference lost direct multi-cause topology")
	}
	_, err = registry.BuildReferenceGraph(context.Background(), []registry.ReferenceCoordinate{coordinate}, referenceResolverFunc(func(_ context.Context, c registry.ReferenceCoordinate) (registry.ReferenceDocument, error) {
		return registry.ReferenceDocument{Coordinate: c, References: []registry.ProviderReference{{Name: "ref", Target: c}}}, nil
	}), registry.GraphLimits{MaxSchemas: 2, MaxDepth: 2, MaxReferences: 2})
	assertPrivateDiagnostic(t, err, nil, registry.ErrReferenceCycle)
	if errors.Unwrap(err) != registry.ErrReferenceCycle {
		t.Fatal("graph cycle lost sentinel wrapper")
	}
}

type privacyCodec struct{ cause error }

func (c privacyCodec) Encode(context.Context, registry.Schema, any) ([]byte, error) {
	return nil, c.cause
}
func (c privacyCodec) Decode(context.Context, registry.Schema, []byte, any) error { return c.cause }

type privacyFramer struct{ cause error }

func (f privacyFramer) Frame(context.Context, registry.ProviderID, []byte) ([]byte, error) {
	return nil, f.cause
}
func (f privacyFramer) Unframe(context.Context, []byte) (registry.ProviderID, []byte, error) {
	return registry.ProviderID{}, nil, f.cause
}

func TestPrivacyCodecCollaborators(t *testing.T) {
	cause := &privacyCause{}
	schema, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"string"`)}, canonicalizerFunc(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil }))
	if err != nil {
		t.Fatal(err)
	}
	id := registry.ProviderID{Provider: "test", Value: "1"}
	codec, err := registry.NewCodecIntegration(privacyCodec{cause}, privacyFramer{cause}, registry.CodecLimits{MaxPayloadBytes: 32, MaxFrameBytes: 32})
	if err != nil {
		t.Fatal(err)
	}
	_, err = codec.Encode(context.Background(), schema, id, nil)
	assertPrivateDiagnostic(t, err, cause, nil)
	if errors.Unwrap(err) != cause {
		t.Fatal("encode lost direct cause")
	}
	_, err = codec.Parse(context.Background(), []byte("frame"))
	assertPrivateDiagnostic(t, err, cause, nil)
	if errors.Unwrap(err) != cause {
		t.Fatal("parse lost direct cause")
	}
	err = codec.Decode(context.Background(), schema, registry.WireMessage{ID: id}, nil)
	assertPrivateDiagnostic(t, err, cause, nil)
	if errors.Unwrap(err) != cause {
		t.Fatal("decode lost direct cause")
	}
	codec, err = registry.NewCodecIntegration(privacyCodec{}, privacyFramer{cause}, registry.CodecLimits{MaxPayloadBytes: 32, MaxFrameBytes: 32})
	if err != nil {
		t.Fatal(err)
	}
	_, err = codec.Encode(context.Background(), schema, id, nil)
	assertPrivateDiagnostic(t, err, cause, nil)
	if errors.Unwrap(err) != cause {
		t.Fatal("frame lost direct cause")
	}
}

func TestPrivacyClientProvider(t *testing.T) {
	cause := &privacyCause{}
	provider := &providerStub{capabilities: registry.Capabilities{Provider: "test", Lookups: []registry.LookupKind{registry.LookupLatest}}, resolve: func(context.Context, registry.Lookup) (registry.ResolveResult, error) {
		return registry.ResolveResult{}, cause
	}}
	client, err := registry.NewClient(provider, registry.Limits{MaxSchemaBytes: 32, MaxListResults: 2, MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Resolve(context.Background(), registry.Latest(registry.Subject{Name: "subject"}))
	assertPrivateDiagnostic(t, err, cause, nil)
}

func TestPrivacyCacheResolver(t *testing.T) {
	cause := &privacyCause{}
	cache, err := registry.NewResolveCache(resolverFunc(func(context.Context, registry.Lookup) (registry.ResolveResult, error) {
		return registry.ResolveResult{}, cause
	}), registry.ResolveCacheConfig{MaxEntries: 2, MaxConcurrent: 1, FreshFor: time.Second, NegativeFor: time.Second, Clock: &testClock{now: time.Unix(1, 0)}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cache.Resolve(context.Background(), registry.Latest(registry.Subject{Name: "subject"}), registry.FailClosed)
	assertPrivateDiagnostic(t, err, cause, nil)
}

func TestPrivacyBundleMissingReferenceIdentity(t *testing.T) {
	canonical := canonicalizerFunc(func(_ context.Context, d registry.Definition) ([]byte, error) { return d.Content, nil })
	leaf, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"string"`)}, canonical)
	if err != nil {
		t.Fatal(err)
	}
	root, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"bytes"`), References: []registry.Reference{{Name: "ref", Fingerprint: leaf.Fingerprint()}}}, canonical)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.NewBundle(root, nil, registry.GraphLimits{MaxSchemas: 2, MaxDepth: 2, MaxReferences: 2}, registry.Provenance{Source: "application", Revision: "1"})
	if !errors.Is(err, registry.ErrReferenceMissing) || strings.Contains(fmt.Sprint(err), leaf.Fingerprint().String()) {
		t.Fatalf("private bundle reference diagnostic: %v", err)
	}
	if errors.Unwrap(err) != registry.ErrReferenceMissing {
		t.Fatal("bundle missing reference lost sentinel wrapper")
	}
}
