package schemaregistry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPrivacyPassthroughTopology(t *testing.T) {
	leaf := errors.New("application-private-detail")
	for _, cause := range []error{leaf, fmt.Errorf("application-private-detail: %w", leaf), errors.Join(leaf, ErrUnavailable)} {
		provider := &basicProvider{capabilities: Capabilities{Provider: "test", Lookups: []LookupKind{LookupByProviderID}}, resolve: func(context.Context, Lookup) (ResolveResult, error) { return ResolveResult{}, cause }}
		client, err := NewClient(provider, validClientLimits())
		if err != nil {
			t.Fatal(err)
		}
		lookup := ByProviderID(ProviderID{Provider: "test", Value: "1"})
		cache, err := NewResolveCache(resolverFunction(provider.Resolve), validCacheConfig(&manualClock{now: time.Unix(1, 0)}))
		if err != nil {
			t.Fatal(err)
		}
		for _, run := range []func() error{func() error { _, err := client.Resolve(t.Context(), lookup); return err }, func() error { _, err := cache.Resolve(t.Context(), lookup, FailClosed); return err }} {
			err := run()
			if err == nil || strings.Contains(err.Error(), "application-private-detail") || !errors.Is(err, cause) {
				t.Fatal("passthrough privacy/identity lost")
			}
			if original, ok := cause.(interface{ Unwrap() []error }); ok {
				actual, ok := err.(interface{ Unwrap() []error })
				if !ok || len(actual.Unwrap()) != len(original.Unwrap()) {
					t.Fatal("passthrough multi-cause topology lost")
				}
				for i, child := range original.Unwrap() {
					if actual.Unwrap()[i] != child {
						t.Fatal("passthrough multi-cause identity lost")
					}
				}
			} else if errors.Unwrap(err) != errors.Unwrap(cause) {
				t.Fatal("passthrough immediate topology lost")
			}
		}
	}
}

func TestPrivacyBareContextCompatibility(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(cause.Error(), func(t *testing.T) {
			schema := internalSchema(t, FormatAvro, `"string"`, nil)
			subject := Subject{Name: "subject"}
			provider := &adminProvider{basicProvider: &basicProvider{
				capabilities: Capabilities{Provider: "test", Formats: []Format{FormatAvro}, Lookups: []LookupKind{LookupByProviderID}, CompatibilityModes: []CompatibilityMode{CompatibilityBackward}, NumericVersions: true, BoundedListing: true, SoftDelete: true},
				register:     func(context.Context, RegisterRequest) (RegisterResult, error) { return RegisterResult{}, cause },
				resolve:      func(context.Context, Lookup) (ResolveResult, error) { return ResolveResult{}, cause },
				compatibility: func(context.Context, CompatibilityRequest) (CompatibilityResult, error) {
					return CompatibilityResult{}, cause
				},
			}, list: func(context.Context, ListRequest) (ListPage, error) { return ListPage{}, cause }, delete: func(context.Context, DeleteRequest) (DeleteResult, error) { return DeleteResult{}, cause }}
			client, err := NewClient(provider, validClientLimits())
			if err != nil {
				t.Fatal(err)
			}
			operations := []struct {
				name string
				run  func() error
			}{
				{"register", func() error {
					_, err := client.Register(t.Context(), RegisterRequest{Subject: subject, Schema: schema})
					return err
				}},
				{"resolve", func() error {
					_, err := client.Resolve(t.Context(), ByProviderID(ProviderID{Provider: "test", Value: "1"}))
					return err
				}},
				{"compatibility", func() error {
					_, err := client.CheckCompatibility(t.Context(), CompatibilityRequest{Subject: subject, Candidate: schema, Mode: CompatibilityBackward})
					return err
				}},
				{"list", func() error { _, err := client.List(t.Context(), ListRequest{Limit: 1}); return err }},
				{"delete", func() error {
					_, err := client.Delete(t.Context(), DeleteRequest{Subject: subject, Version: Version{Number: 1}, Policy: DeletionPolicy{Mode: DeleteSoft, ExpectedFingerprint: schema.Fingerprint()}})
					return err
				}},
			}
			for _, operation := range operations {
				t.Run(operation.name, func(t *testing.T) {
					if err := operation.run(); err != cause {
						t.Fatalf("bare context identity lost: %T", err)
					}
				})
			}
			cache, err := NewResolveCache(resolverFunction(func(context.Context, Lookup) (ResolveResult, error) { return ResolveResult{}, cause }), validCacheConfig(&manualClock{now: time.Unix(1, 0)}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := cache.Resolve(t.Context(), ByProviderID(ProviderID{Provider: "test", Value: "1"}), FailClosed); err != cause {
				t.Fatalf("cache bare context identity lost: %T", err)
			}
			integration, err := NewCodecIntegration(compatibilityCodec{cause}, compatibilityFramer{}, CodecLimits{MaxPayloadBytes: 32, MaxFrameBytes: 32})
			if err != nil {
				t.Fatal(err)
			}
			_, err = integration.Encode(t.Context(), schema, ProviderID{Provider: "test", Value: "1"}, nil)
			if err == cause || errors.Unwrap(err) != cause {
				t.Fatal("formerly wrapped codec context topology changed")
			}
		})
	}
}

type compatibilityCodec struct{ cause error }

func (c compatibilityCodec) Encode(context.Context, Schema, any) ([]byte, error) { return nil, c.cause }
func (c compatibilityCodec) Decode(context.Context, Schema, []byte, any) error   { return c.cause }

type compatibilityFramer struct{}

func (compatibilityFramer) Frame(context.Context, ProviderID, []byte) ([]byte, error) {
	return nil, nil
}
func (compatibilityFramer) Unframe(context.Context, []byte) (ProviderID, []byte, error) {
	return ProviderID{}, nil, nil
}
