package schemaregistry_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
)

func TestClientBoundsRegistrationOwnersBeforeProviderAdmission(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	var calls atomic.Int32
	provider := &providerStub{
		capabilities: schemaregistry.Capabilities{Provider: "test", Formats: []schemaregistry.Format{schemaregistry.FormatAvro}},
		register: func(context.Context, schemaregistry.RegisterRequest) (schemaregistry.RegisterResult, error) {
			if calls.Add(1) == 1 {
				close(started)
				<-release
			}
			return schemaregistry.RegisterResult{Outcome: schemaregistry.RegistrationExisting, ID: schemaregistry.ProviderID{Provider: "test", Value: "1"}}, nil
		},
	}
	client, err := schemaregistry.NewClient(provider, schemaregistry.Limits{MaxSchemaBytes: 1024, MaxListResults: 10, MaxConcurrent: 1})
	if err != nil {
		t.Fatal(err)
	}
	request := schemaregistry.RegisterRequest{Subject: schemaregistry.Subject{Name: "first"}, Schema: compileAvroString(t)}
	first := make(chan error, 1)
	completed := false
	go func() { _, err := client.Register(context.Background(), request); first <- err }()
	defer func() {
		unblock.Do(func() { close(release) })
		if !completed {
			if err := waitCacheTestValue(t, first, "registration owner completion"); err != nil {
				t.Errorf("first Register() = %v", err)
			}
		}
	}()
	waitCacheTestValue(t, started, "provider admission")
	joinCtx, cancelJoin := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelJoin()
	if _, err := client.Register(joinCtx, request); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("canceled same-key Register() = %v", err)
	}
	request.Subject.Name = "second"
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	result, err := client.Register(ctx, request)
	if !errors.Is(err, schemaregistry.ErrLimitExceeded) || result != (schemaregistry.RegisterResult{}) || calls.Load() != 1 {
		t.Fatalf("Register(at capacity) = (%+v, %v), provider calls=%d", result, err, calls.Load())
	}
	unblock.Do(func() { close(release) })
	err = waitCacheTestValue(t, first, "released registration owner")
	completed = true
	if err != nil {
		t.Fatal(err)
	}
	result, err = client.Register(context.Background(), request)
	if err != nil || result.Outcome != schemaregistry.RegistrationExisting || calls.Load() != 2 {
		t.Fatalf("Register(after completion) = (%+v, %v), provider calls=%d", result, err, calls.Load())
	}
}

func TestResolveCacheBoundsActiveOwnersIncludingDetachedGenerations(t *testing.T) {
	for _, detached := range []bool{false, true} {
		name := "distinct selector"
		if detached {
			name = "detached generation"
		}
		t.Run(name, func(t *testing.T) {
			started, release := make(chan struct{}), make(chan struct{})
			var unblock sync.Once
			var calls atomic.Int32
			schema := compileAvroString(t)
			cache, err := schemaregistry.NewResolveCache(resolverFunc(func(_ context.Context, lookup schemaregistry.Lookup) (schemaregistry.ResolveResult, error) {
				if calls.Add(1) == 1 {
					close(started)
					<-release
				}
				return schemaregistry.ResolveResult{ID: lookup.ProviderID(), Schema: schema, Lifecycle: schemaregistry.LifecycleAvailable}, nil
			}), schemaregistry.ResolveCacheConfig{MaxEntries: 2, MaxConcurrent: 1, FreshFor: time.Minute, NegativeFor: time.Minute, Clock: &testClock{now: time.Unix(100, 0)}})
			if err != nil {
				t.Fatal(err)
			}
			lookup := schemaregistry.ByProviderID(schemaregistry.ProviderID{Provider: "test", Value: "1"})
			first := make(chan error, 1)
			completed := false
			go func() { _, err := cache.Resolve(context.Background(), lookup, schemaregistry.FailClosed); first <- err }()
			defer func() {
				unblock.Do(func() { close(release) })
				if !completed {
					if err := waitCacheTestValue(t, first, "cache owner completion"); err != nil {
						t.Errorf("first Resolve() = %v", err)
					}
				}
			}()
			waitCacheTestValue(t, started, "resolver admission")
			joinCtx, cancelJoin := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancelJoin()
			if _, err := cache.Resolve(joinCtx, lookup, schemaregistry.FailClosed); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("same-key waiter did not retain independent cancellation: %v", err)
			}
			second := schemaregistry.ByProviderID(schemaregistry.ProviderID{Provider: "test", Value: "2"})
			if detached {
				if err := cache.Invalidate(lookup); err != nil {
					t.Fatal(err)
				}
				second = lookup
			}
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()
			result, err := cache.Resolve(ctx, second, schemaregistry.AllowStale)
			if !errors.Is(err, schemaregistry.ErrLimitExceeded) || !reflect.DeepEqual(result, schemaregistry.CacheResolution{}) || calls.Load() != 1 {
				t.Fatalf("Resolve(at capacity) = (%+v, %v), resolver calls=%d", result, err, calls.Load())
			}
			unblock.Do(func() { close(release) })
			err = waitCacheTestValue(t, first, "released cache owner")
			completed = true
			if err != nil {
				t.Fatal(err)
			}
			if detached {
				if _, err := cache.Resolve(context.Background(), lookup, schemaregistry.CacheOnly); !errors.Is(err, schemaregistry.ErrOfflineMiss) {
					t.Fatalf("detached owner populated cache: %v", err)
				}
			}
			result, err = cache.Resolve(context.Background(), second, schemaregistry.FailClosed)
			if err != nil || result.State != schemaregistry.CacheLoaded || result.Result.ID != second.ProviderID() || calls.Load() != 2 {
				t.Fatalf("Resolve(after completion) = (%+v, %v), resolver calls=%d", result, err, calls.Load())
			}
		})
	}
}
