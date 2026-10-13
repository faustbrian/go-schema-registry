package schemaregistry_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
)

// Signal the waiter's entry into its select without scheduler sleeps.
type panicWaiterContext struct {
	context.Context
	joined chan struct{}
	once   sync.Once
}

func (ctx *panicWaiterContext) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.joined) })
	return ctx.Context.Done()
}

func recoveredRegistryPanic(operation func()) (value any) {
	defer func() { value = recover() }()
	operation()
	return nil
}

func TestPanicReleasesOwnedFlightAndAdmission(t *testing.T) {
	for _, variant := range []string{"registration", "cache", "detached-cache"} {
		t.Run(variant, func(t *testing.T) {
			schema := compileAvroString(t)
			panicValue := new(int)
			started, release := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			upstream := func() {
				if calls.Add(1) == 1 {
					close(started)
					<-release
					panic(panicValue)
				}
			}
			var run func(context.Context) error
			invalidate := func() {}
			if variant == "registration" {
				client, err := schemaregistry.NewClient(&providerStub{
					capabilities: schemaregistry.Capabilities{Provider: "test", Formats: []schemaregistry.Format{schemaregistry.FormatAvro}},
					register: func(context.Context, schemaregistry.RegisterRequest) (schemaregistry.RegisterResult, error) {
						upstream()
						return schemaregistry.RegisterResult{Outcome: schemaregistry.RegistrationExisting, ID: schemaregistry.ProviderID{Provider: "test", Value: "1"}}, nil
					},
				}, schemaregistry.Limits{MaxSchemaBytes: 1024, MaxListResults: 10, MaxConcurrent: 1})
				if err != nil {
					t.Fatal(err)
				}
				request := schemaregistry.RegisterRequest{Subject: schemaregistry.Subject{Name: "subject"}, Schema: schema}
				run = func(ctx context.Context) error {
					result, err := client.Register(ctx, request)
					if errors.Is(err, schemaregistry.ErrUnavailable) && result.Outcome != schemaregistry.RegistrationUnknown {
						t.Errorf("aborted registration claimed known outcome: %s", result.Outcome)
					}
					return err
				}
			} else {
				lookup := schemaregistry.ByFingerprint(schema.Fingerprint())
				cache, err := schemaregistry.NewResolveCache(resolverFunc(func(context.Context, schemaregistry.Lookup) (schemaregistry.ResolveResult, error) {
					upstream()
					return schemaregistry.ResolveResult{Schema: schema, ID: schemaregistry.ProviderID{Provider: "test", Value: "1"}, Lifecycle: schemaregistry.LifecycleAvailable}, nil
				}), schemaregistry.ResolveCacheConfig{MaxEntries: 1, MaxConcurrent: 1, FreshFor: time.Minute, NegativeFor: time.Minute, Clock: &testClock{now: time.Unix(0, 0)}})
				if err != nil {
					t.Fatal(err)
				}
				run = func(ctx context.Context) error {
					_, err := cache.Resolve(ctx, lookup, schemaregistry.FailClosed)
					return err
				}
				if variant == "detached-cache" {
					invalidate = func() {
						if err := cache.Invalidate(lookup); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			leader := make(chan any, 1)
			go func() { leader <- recoveredRegistryPanic(func() { _ = run(t.Context()) }) }()
			waitCacheTestValue(t, started, "panicking owner admission")
			base, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			waiterCtx := &panicWaiterContext{Context: base, joined: make(chan struct{})}
			waiter := make(chan error, 1)
			go func() { waiter <- run(waiterCtx) }()
			waitCacheTestValue(t, waiterCtx.joined, "same-flight waiter select")
			invalidate()
			close(release)
			if got := waitCacheTestValue(t, leader, "panicking owner return"); got != panicValue {
				t.Fatal("trusted panic value did not propagate unchanged")
			}
			if err := waitCacheTestValue(t, waiter, "aborted flight waiter"); !errors.Is(err, schemaregistry.ErrUnavailable) {
				t.Fatalf("waiter did not receive safe aborted-flight failure: %v", err)
			}
			if err := run(base); err != nil {
				t.Fatalf("panic poisoned subsequent admission: %v", err)
			}
		})
	}
}

type loadPanicClock struct {
	calls atomic.Int32
	value any
}

func (clock *loadPanicClock) Now() time.Time {
	if clock.calls.Add(1) == 2 {
		panic(clock.value)
	}
	return time.Unix(0, 0)
}

func TestLoadClockPanicReleasesFlightAdmission(t *testing.T) {
	schema := compileAvroString(t)
	panicValue := new(int)
	cache, err := schemaregistry.NewResolveCache(resolverFunc(func(context.Context, schemaregistry.Lookup) (schemaregistry.ResolveResult, error) {
		return schemaregistry.ResolveResult{Schema: schema, ID: schemaregistry.ProviderID{Provider: "test", Value: "1"}, Lifecycle: schemaregistry.LifecycleAvailable}, nil
	}), schemaregistry.ResolveCacheConfig{MaxEntries: 1, MaxConcurrent: 1, FreshFor: time.Minute, NegativeFor: time.Minute, Clock: &loadPanicClock{value: panicValue}})
	if err != nil {
		t.Fatal(err)
	}
	lookup := schemaregistry.ByFingerprint(schema.Fingerprint())
	if got := recoveredRegistryPanic(func() { _, _ = cache.Resolve(t.Context(), lookup, schemaregistry.FailClosed) }); got != panicValue {
		t.Fatal("load Clock panic did not propagate unchanged")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
	defer cancel()
	if _, err := cache.Resolve(ctx, lookup, schemaregistry.FailClosed); err != nil {
		t.Fatalf("load Clock panic poisoned flight admission: %v", err)
	}
}

func TestListingPanicReleasesClientAdmission(t *testing.T) {
	panicValue := new(int)
	var calls atomic.Int32
	provider := &administrativeProviderStub{
		providerStub: providerStub{capabilities: schemaregistry.Capabilities{Provider: "test", BoundedListing: true}},
		list: func(context.Context, schemaregistry.ListRequest) (schemaregistry.ListPage, error) {
			if calls.Add(1) == 1 {
				panic(panicValue)
			}
			return schemaregistry.ListPage{}, nil
		},
	}
	client, err := schemaregistry.NewClient(provider, schemaregistry.Limits{MaxSchemaBytes: 1024, MaxListResults: 10, MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	request := schemaregistry.ListRequest{Limit: 1}
	if got := recoveredRegistryPanic(func() { _, _ = client.List(t.Context(), request) }); got != panicValue {
		t.Fatal("listing panic did not propagate unchanged")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
	defer cancel()
	if _, err := client.List(ctx, request); err != nil {
		t.Fatalf("listing panic poisoned client admission: %v", err)
	}
}
