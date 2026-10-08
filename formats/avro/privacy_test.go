package avro_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	registry "github.com/faustbrian/go-schema-registry/v3"
	"github.com/faustbrian/go-schema-registry/v3/formats/avro"
)

func TestPrivacyCanonicalizerDiagnostic(t *testing.T) {
	_, err := avro.New(1024).Canonicalize(context.Background(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`"application-private-detail"`)})
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatalf("private Avro diagnostic: %v", err)
	}
	cause := errors.Unwrap(err)
	if cause == nil || !errors.Is(err, cause) || !strings.Contains(cause.Error(), "application-private-detail") {
		t.Fatal("explicit Avro cause inspection lost the original parser diagnostic")
	}
}

func TestPrivacyMalformedJSONPreservesSyntaxCause(t *testing.T) {
	_, err := avro.New(1024).Canonicalize(t.Context(), registry.Definition{Format: registry.FormatAvro, Content: []byte(`{"application-private-detail":]}`)})
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) || errors.Unwrap(err) != syntax || !errors.Is(err, syntax) || syntax.Offset <= 0 {
		t.Fatal("malformed Avro JSON lost its direct structured syntax cause")
	}
	if strings.Contains(err.Error(), "application-private-detail") {
		t.Fatal("default JSON diagnostic exposes schema detail")
	}
}
