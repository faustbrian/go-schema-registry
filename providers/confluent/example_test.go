package confluent_test

import (
	"fmt"
	"net/http"
	"time"

	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	"github.com/faustbrian/go-schema-registry/providers/confluent"
)

func ExampleNew() {
	provider, err := confluent.New(confluent.Config{
		Endpoint:         "https://registry.example.com",
		Scope:            "production-cluster",
		Transport:        http.DefaultTransport,
		RequestTimeout:   5 * time.Second,
		MaxResponseBytes: 1 << 20,
		MaxAttempts:      3,
		MaxConcurrent:    8,
		RetryDelay:       100 * time.Millisecond,
		ReferenceLimits: schemaregistry.GraphLimits{
			MaxSchemas: 32, MaxDepth: 8, MaxReferences: 64,
		},
		Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{
			schemaregistry.FormatAvro: avro.New(1 << 20),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(provider.Capabilities().Provider)
	// Output: confluent
}
