// Package schemaregistry provides explicit schema compilation, scoped provider
// clients, cache/bundle/graph resolution and codec/framer composition. Start with
// Compile and an explicit format canonicalizer, then NewClient/NewResolveCache or
// NewBundle as needed. Schema/Bundle values own copied data; clients/cache own
// synchronized local state but no background goroutines or collaborator shutdown.
//
// Portable identity is this package's SHA-256 profile, not a provider ID,
// authenticity proof or universal compatibility relation. Local-only
// canonicalizers/codecs are collaborator obligations, not an I/O sandbox. Context
// cannot forcibly interrupt callbacks, and per-owner limits do not bound total
// process memory, waiters or deployment demand. Root /v3 and Confluent /v3 types
// differ from Glue's retained /v2 cohort. Exact language/source versions and
// SCHEMA-REG-DEC-001..004 are in specification/README.md and spec/sources.lock.json.
package schemaregistry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"slices"
)

var (
	// ErrInvalidSchema marks malformed or otherwise invalid schema input.
	ErrInvalidSchema = errors.New("schema registry: invalid schema")
	// ErrUnsupportedFormat marks a schema format for which no canonical
	// implementation was supplied.
	ErrUnsupportedFormat = errors.New("schema registry: unsupported format")
)

// CompileLimits bound hostile schema definitions before and after format
// canonicalization.
type CompileLimits struct {
	// MaxSchemaBytes caps content and, independently, total reference text and
	// metadata text before copying; it is not a combined allocation budget.
	MaxSchemaBytes int
	// MaxCanonicalBytes caps returned canonical bytes after collaborator work.
	MaxCanonicalBytes int
	// MaxReferences caps named reference count; each fingerprint must be nonzero.
	MaxReferences int
	// MaxMetadata caps entry count; all limits must be positive, not zero-defaulted.
	MaxMetadata int
}

// DefaultCompileLimits returns conservative portable bounds. Provider adapters
// may impose stricter service limits. Content/canonical bytes and each reference/
// metadata text aggregate default to 1MiB; reference and metadata counts to 256.
func DefaultCompileLimits() CompileLimits {
	return CompileLimits{
		MaxSchemaBytes:    1 << 20,
		MaxCanonicalBytes: 1 << 20,
		MaxReferences:     256,
		MaxMetadata:       256,
	}
}

// Format identifies one schema language. Format is portable; provider-specific
// names are mapped only by adapters.
type Format string

const (
	// FormatJSONSchema identifies JSON Schema documents.
	FormatJSONSchema Format = "json-schema"
	// FormatAvro identifies Apache Avro schemas.
	FormatAvro Format = "avro"
	// FormatProtobuf identifies Protocol Buffers schemas.
	FormatProtobuf Format = "protobuf"
)

// Reference identifies another versioned schema used by a definition.
type Reference struct {
	// Name is a nonempty unique binding included in portable identity.
	Name string
	// Subject is a provider coordinate, excluded from the portable hash.
	Subject string
	// Version is a numeric provider coordinate, excluded from the portable hash.
	Version uint64
	// Fingerprint is the nonzero portable dependency identity; Compile does not
	// resolve the dependency or verify that it matches Subject/Version.
	Fingerprint Fingerprint
}

// Definition is caller-owned schema input. Compile copies all mutable fields.
type Definition struct {
	// Format selects an explicitly supplied supported canonicalizer.
	Format Format
	// Content is raw nonempty schema text, copied during Compile.
	Content []byte
	// References is copied and sorted by Name for hashing; caller storage must
	// remain stable while copied. Format adapters may require separate local resources.
	References []Reference
	// Metadata is copied but excluded from portable identity, not redacted or
	// authenticated. Keys/values can retain immutable string backing storage.
	Metadata map[string]string
}

// Canonicalizer validates a definition and returns the canonical bytes used
// for portable identity. Implementations must not contact a registry.
type Canonicalizer interface {
	// Canonicalize must validate the selected language/profile without implicit
	// remote retrieval and cooperate with ctx. Compile supplies an owned copy;
	// direct Canonicalize callers retain input ownership. Compile copies admitted
	// returned bytes. Calls can be concurrent
	// when a canonicalizer is shared; implementations own that safety and bounds.
	Canonicalize(context.Context, Definition) ([]byte, error)
}

// Fingerprint is a portable SHA-256 identity over the format and canonical
// schema. It is never interchangeable with a provider-issued identifier.
type Fingerprint struct {
	sum [sha256.Size]byte
}

// String returns the algorithm-qualified lowercase hexadecimal fingerprint.
func (fingerprint Fingerprint) String() string {
	return "sha256:" + hex.EncodeToString(fingerprint.sum[:])
}

// ParseFingerprint accepts the exact sha256: prefix and 64 hexadecimal digits,
// including uppercase digits; String normalizes them to lowercase. It checks
// syntax only and does not authenticate content or reject the all-zero digest.
func ParseFingerprint(value string) (Fingerprint, error) {
	const prefix = "sha256:"
	if len(value) != len(prefix)+sha256.Size*2 || value[:min(len(value), len(prefix))] != prefix {
		return Fingerprint{}, fmt.Errorf("%w: fingerprint", ErrInvalidSchema)
	}
	decoded, err := hex.DecodeString(value[len(prefix):])
	if err != nil {
		return Fingerprint{}, fmt.Errorf("%w: fingerprint", ErrInvalidSchema)
	}
	var sum [sha256.Size]byte
	copy(sum[:], decoded)
	return Fingerprint{sum: sum}, nil
}

// Schema is an immutable compiled definition and its portable identity.
type Schema struct {
	definition  Definition
	canonical   []byte
	fingerprint Fingerprint
}

// Compile applies DefaultCompileLimits and delegates to the supplied canonicalizer.
// No registry is contacted by this wrapper; local-only behavior is the
// canonicalizer's obligation, not enforced isolation.
func Compile(
	ctx context.Context,
	definition Definition,
	canonicalizer Canonicalizer,
) (Schema, error) {
	return CompileWithLimits(ctx, definition, canonicalizer, DefaultCompileLimits())
}

// CompileWithLimits validates and canonicalizes one schema within explicit
// caller-selected input/result bounds. It copies caller mutable data, sorts named
// references and hashes the domain separator, format, canonical bytes and each
// Name/Fingerprint pair. Subject/version coordinates and metadata are excluded.
// The canonicalizer receives a second copy; output caps follow its allocation,
// and no final cancellation check, panic containment or callback timeout is added.
// Errors expose classified sentinels with fixed text; explicit causes can retain
// collaborator diagnostics. See SCHEMA-REG-DEC-004 and portable corpus tests.
func CompileWithLimits(
	ctx context.Context,
	definition Definition,
	canonicalizer Canonicalizer,
	limits CompileLimits,
) (Schema, error) {
	if err := ctx.Err(); err != nil {
		return Schema{}, err
	}
	if limits.MaxSchemaBytes <= 0 || limits.MaxCanonicalBytes <= 0 ||
		limits.MaxReferences <= 0 || limits.MaxMetadata <= 0 {
		return Schema{}, fmt.Errorf("%w: compile limits", ErrInvalidRequest)
	}
	if !definition.Format.valid() {
		return Schema{}, fmt.Errorf("%w: schema format", ErrUnsupportedFormat)
	}
	if len(definition.Content) == 0 {
		return Schema{}, fmt.Errorf("%w: empty content", ErrInvalidSchema)
	}
	if len(definition.Content) > limits.MaxSchemaBytes ||
		len(definition.References) > limits.MaxReferences ||
		len(definition.Metadata) > limits.MaxMetadata {
		return Schema{}, fmt.Errorf("%w: schema definition", ErrLimitExceeded)
	}
	referenceBytes := uint64(0)
	for _, reference := range definition.References {
		var withinLimit bool
		referenceBytes, withinLimit = boundedDefinitionText(referenceBytes, len(reference.Name), limits.MaxSchemaBytes)
		if !withinLimit {
			return Schema{}, fmt.Errorf("%w: schema reference text", ErrLimitExceeded)
		}
		referenceBytes, withinLimit = boundedDefinitionText(referenceBytes, len(reference.Subject), limits.MaxSchemaBytes)
		if !withinLimit {
			return Schema{}, fmt.Errorf("%w: schema reference text", ErrLimitExceeded)
		}
	}
	metadataBytes := uint64(0)
	for key, value := range definition.Metadata {
		var withinLimit bool
		metadataBytes, withinLimit = boundedDefinitionText(metadataBytes, len(key), limits.MaxSchemaBytes)
		if !withinLimit {
			return Schema{}, fmt.Errorf("%w: schema metadata text", ErrLimitExceeded)
		}
		metadataBytes, withinLimit = boundedDefinitionText(metadataBytes, len(value), limits.MaxSchemaBytes)
		if !withinLimit {
			return Schema{}, fmt.Errorf("%w: schema metadata text", ErrLimitExceeded)
		}
	}
	if interfaceIsNil(canonicalizer) {
		return Schema{}, fmt.Errorf("%w: no %s canonicalizer", ErrUnsupportedFormat, definition.Format)
	}
	seenReferences := make(map[string]struct{}, len(definition.References))
	for _, reference := range definition.References {
		if reference.Name == "" || reference.Fingerprint == (Fingerprint{}) {
			return Schema{}, fmt.Errorf("%w: unresolved reference", ErrInvalidSchema)
		}
		if _, exists := seenReferences[reference.Name]; exists {
			return Schema{}, fmt.Errorf("%w: duplicate reference", ErrInvalidSchema)
		}
		seenReferences[reference.Name] = struct{}{}
	}

	owned := cloneDefinition(definition)
	slices.SortFunc(owned.References, compareReferenceNames)
	canonical, err := canonicalizer.Canonicalize(ctx, cloneDefinition(owned))
	if err != nil {
		return Schema{}, privateDiagnostic("schema registry: invalid schema: canonicalization failed", ErrInvalidSchema, err)
	}
	if len(canonical) == 0 {
		return Schema{}, fmt.Errorf("%w: empty canonical form", ErrInvalidSchema)
	}
	if len(canonical) > limits.MaxCanonicalBytes {
		return Schema{}, fmt.Errorf("%w: canonical schema", ErrLimitExceeded)
	}

	// Reference names are sorted, but their fingerprints (not provider coordinates)
	// bind portable dependencies. Keep this historical v1 domain separator stable:
	// changing it changes every identity even though the module is now /v3.
	hash := sha256.New()
	_, _ = hash.Write([]byte("golib.schema-registry/v1\x00"))
	_, _ = hash.Write([]byte(definition.Format))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(canonical)
	for _, reference := range owned.References {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(reference.Name))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(reference.Fingerprint.sum[:])
	}
	var sum [sha256.Size]byte
	copy(sum[:], hash.Sum(nil))

	return Schema{
		definition:  owned,
		canonical:   append([]byte(nil), canonical...),
		fingerprint: Fingerprint{sum: sum},
	}, nil
}

func boundedDefinitionText(current uint64, additional, limit int) (uint64, bool) {
	// #nosec G115 -- All production callers pass nonnegative string lengths; positive limits are validated before this sum.
	total, carry := bits.Add64(current, uint64(additional), 0)
	// #nosec G115 -- All production callers supply a validated positive MaxSchemaBytes limit; carry rejects overflow.
	return total, carry == 0 && total <= uint64(limit)
}

func compareReferenceNames(left, right Reference) int {
	if left.Name < right.Name {
		return -1
	}
	if left.Name > right.Name {
		return 1
	}
	return 0
}

// Definition returns a copy of the schema definition.
func (schema Schema) Definition() Definition { return cloneDefinition(schema.definition) }

// Fingerprint returns the portable schema identity.
func (schema Schema) Fingerprint() Fingerprint { return schema.fingerprint }

// Canonical returns a copy of the validated canonical representation.
func (schema Schema) Canonical() []byte { return append([]byte(nil), schema.canonical...) }

func (format Format) valid() bool {
	switch format {
	case FormatJSONSchema, FormatAvro, FormatProtobuf:
		return true
	default:
		return false
	}
}

func cloneDefinition(definition Definition) Definition {
	clone := Definition{
		Format:  definition.Format,
		Content: append([]byte(nil), definition.Content...),
	}
	if definition.References != nil {
		clone.References = append([]Reference(nil), definition.References...)
	}
	if definition.Metadata != nil {
		clone.Metadata = make(map[string]string, len(definition.Metadata))
		for key, value := range definition.Metadata {
			clone.Metadata[key] = value
		}
	}
	return clone
}
