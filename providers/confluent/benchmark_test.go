package confluent_test

import (
	"context"
	"testing"

	"github.com/faustbrian/go-schema-registry/providers/confluent/v3"
	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
	"github.com/twmb/franz-go/pkg/sr"
)

// BenchmarkClassicFrame measures allocation/framing of a fixed 1024-byte
// payload and scoped ID, excluding setup, business serialization and network.
func BenchmarkClassicFrame(b *testing.B) {
	framer, err := confluent.NewClassicFramer("bench", 4096)
	if err != nil {
		b.Fatalf("NewClassicFramer() error = %v", err)
	}
	id := schemaregistry.ProviderID{Provider: confluent.ProviderName, Scope: "bench", Value: "1"}
	payload := make([]byte, 1024)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := framer.Frame(context.Background(), id, payload); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFranzGoClassicFrame measures the same fixed payload/header with
// franz-go sr1.8.0; compare identical corpus/toolchain and allocation boundaries.
func BenchmarkFranzGoClassicFrame(b *testing.B) {
	payload := make([]byte, 1024)
	var header sr.ConfluentHeader
	b.ReportAllocs()
	for b.Loop() {
		frame, err := header.AppendEncode(nil, 1, nil)
		if err != nil {
			b.Fatal(err)
		}
		frame = append(frame, payload...)
		if len(frame) != len(payload)+5 {
			b.Fatal("unexpected frame size")
		}
	}
}

// BenchmarkProtobufFrame measures fixed 1024-byte payload framing with [0]
// indexes, excluding setup and descriptor/business-codec/provider work.
func BenchmarkProtobufFrame(b *testing.B) {
	payload := make([]byte, 1024)
	framer, err := confluent.NewProtobufFramer("bench", len(payload), 1)
	if err != nil {
		b.Fatal(err)
	}
	id := schemaregistry.ProviderID{Provider: confluent.ProviderName, Scope: "bench", Value: "1"}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := framer.FrameMessage(context.Background(), id, []int{0}, payload); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFranzGoProtobufFrame measures the equivalent [0]-path header and
// fixed payload with the selected peer; it is not registry latency evidence.
func BenchmarkFranzGoProtobufFrame(b *testing.B) {
	payload := make([]byte, 1024)
	var header sr.ConfluentHeader
	b.ReportAllocs()
	for b.Loop() {
		frame, err := header.AppendEncode(nil, 1, []int{0})
		if err != nil {
			b.Fatal(err)
		}
		frame = append(frame, payload...)
		if len(frame) != len(payload)+6 {
			b.Fatal("unexpected frame size")
		}
	}
}
