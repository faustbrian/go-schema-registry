package jsonschema_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	registry "github.com/faustbrian/go-schema-registry"
	adapterjson "github.com/faustbrian/go-schema-registry/formats/jsonschema"
)

type privateMarshaler struct{ cause error }

func (m privateMarshaler) MarshalJSON() ([]byte, error) { return nil, m.cause }

func TestPrivacyPayloadDiagnostic(t *testing.T) {
	adapter, err := adapterjson.New(adapterjson.Config{MaxSchemaBytes: 1024, MaxTotalSchemaBytes: 2048, MaxPayloadBytes: 128, MaxResources: 2})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := registry.Compile(context.Background(), registry.Definition{Format: registry.FormatJSONSchema, Content: []byte(`{}`)}, adapter)
	if err != nil {
		t.Fatal(err)
	}
	cause := errors.New("application-private-detail")
	_, err = adapter.Encode(context.Background(), schema, privateMarshaler{cause})
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Errorf("private JSON payload diagnostic: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("original marshal cause lost")
	}
	var marshal *json.MarshalerError
	if !errors.As(err, &marshal) || errors.Unwrap(err) != marshal || errors.Unwrap(marshal) != cause {
		t.Fatal("marshal failure lost its direct structured cause topology")
	}
}
