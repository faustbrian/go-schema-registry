package protobuf_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	registry "github.com/faustbrian/go-schema-registry/v2"
	"github.com/faustbrian/go-schema-registry/v2/formats/protobuf"
)

func TestPrivacyCanonicalizerDiagnostic(t *testing.T) {
	c, err := protobuf.New(protobuf.Config{Filename: "application-private-detail.proto", MaxSchemaBytes: 1024, MaxImports: 2})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Canonicalize(context.Background(), registry.Definition{Format: registry.FormatProtobuf, Content: []byte(`syntax = "proto3"; message Test { Missing field = 1; }`)})
	if err == nil || strings.Contains(err.Error(), "application-private-detail") {
		t.Fatalf("private Protobuf diagnostic: %v", err)
	}
	cause := errors.Unwrap(err)
	if cause == nil || !errors.Is(err, cause) || !strings.Contains(cause.Error(), "application-private-detail.proto") {
		t.Fatal("explicit Protobuf cause inspection lost the original source diagnostic")
	}
}
