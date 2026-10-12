package schemaregistry

import (
	"context"
	"fmt"
)

// ValueCodec performs business serialization against an already-resolved
// schema. Implementations must not access a registry.
type ValueCodec interface {
	// Encode must apply the resolved schema without hidden registry discovery;
	// return bytes or error and honor ctx. Arbitrary business values/callbacks
	// may allocate or mutate externally; bounds/cooperation belong to the codec.
	Encode(context.Context, Schema, any) ([]byte, error)
	// Decode must apply schema semantics to payload and write the caller target;
	// target may be partially mutated on error. No automatic rollback is promised.
	Decode(context.Context, Schema, []byte, any) error
}

// Framer owns one explicitly versioned provider wire format. It receives and
// returns opaque provider IDs, never portable fingerprints.
type Framer interface {
	// Frame must preserve the selected version/scoped identity and payload;
	// cooperate with ctx and make returned-byte ownership explicit.
	Frame(context.Context, ProviderID, []byte) ([]byte, error)
	// Unframe must validate the declared wire version and scoped ID; it may
	// return borrowed payload bytes because Parse copies admitted results.
	Unframe(context.Context, []byte) (ProviderID, []byte, error)
}

// CodecLimits bound both business payloads and complete framed messages.
type CodecLimits struct {
	// MaxPayloadBytes is positive; encoding checks it after codec allocation,
	// decoding checks it before passing a copied payload to the codec.
	MaxPayloadBytes int
	// MaxFrameBytes must be at least MaxPayloadBytes; parsed input is checked
	// before Unframe, encoded output after framing allocation.
	MaxFrameBytes int
}

// WireMessage is parsed framing metadata plus an owned payload copy. Callers
// explicitly resolve ID before invoking Decode.
type WireMessage struct {
	// ID is opaque framing evidence; Decode does not verify its relation to Schema.
	ID ProviderID
	// Payload is mutable owned bytes after Parse; direct callers must keep it
	// stable while Decode makes its defensive copy.
	Payload []byte
}

// CodecIntegration composes a business codec and provider framer without a
// built-in registry dependency. Borrowed collaborators must uphold the no-hidden
// registry contract; no sandbox, panic recovery or callback preemption is added.
// Concurrent reuse depends on codec/framer safety. It owns no Close lifecycle.
type CodecIntegration struct {
	codec  ValueCodec
	framer Framer
	limits CodecLimits
}

// NewCodecIntegration validates explicit codec and framing bounds.
func NewCodecIntegration(
	codec ValueCodec,
	framer Framer,
	limits CodecLimits,
) (*CodecIntegration, error) {
	if interfaceIsNil(codec) || interfaceIsNil(framer) {
		return nil, fmt.Errorf("%w: codec and framer are required", ErrInvalidRequest)
	}
	if limits.MaxPayloadBytes <= 0 || limits.MaxFrameBytes < max(1, limits.MaxPayloadBytes) {
		return nil, fmt.Errorf("%w: codec limits", ErrInvalidRequest)
	}
	return &CodecIntegration{codec: codec, framer: framer, limits: limits}, nil
}

// Encode serializes against an explicit schema, checks payload bounds, and
// frames with an explicit provider ID.
func (integration *CodecIntegration) Encode(
	ctx context.Context,
	schema Schema,
	id ProviderID,
	value any,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if schema.Fingerprint() == (Fingerprint{}) || id.Provider == "" || id.Value == "" {
		return nil, fmt.Errorf("%w: schema and provider ID are required", ErrInvalidRequest)
	}
	payload, err := integration.codec.Encode(ctx, schema, value)
	if err != nil {
		return nil, privateDiagnostic("schema registry: encode value failed", err)
	}
	if len(payload) > integration.limits.MaxPayloadBytes {
		return nil, fmt.Errorf("%w: payload bytes", ErrLimitExceeded)
	}
	framed, err := integration.framer.Frame(ctx, id, payload)
	if err != nil {
		return nil, privateDiagnostic("schema registry: frame payload failed", err)
	}
	if len(framed) > integration.limits.MaxFrameBytes {
		return nil, fmt.Errorf("%w: frame bytes", ErrLimitExceeded)
	}
	return append([]byte(nil), framed...), nil
}

// Parse validates the complete frame bound before delegating to the explicit
// provider framer. It never resolves the returned ID.
func (integration *CodecIntegration) Parse(
	ctx context.Context,
	framed []byte,
) (WireMessage, error) {
	if err := ctx.Err(); err != nil {
		return WireMessage{}, err
	}
	if len(framed) > integration.limits.MaxFrameBytes {
		return WireMessage{}, fmt.Errorf("%w: frame bytes", ErrLimitExceeded)
	}
	id, payload, err := integration.framer.Unframe(ctx, framed)
	if err != nil {
		return WireMessage{}, privateDiagnostic("schema registry: unframe payload failed", err)
	}
	if id.Provider == "" || id.Value == "" {
		return WireMessage{}, fmt.Errorf("%w: provider ID", ErrInvalidSchema)
	}
	if len(payload) > integration.limits.MaxPayloadBytes {
		return WireMessage{}, fmt.Errorf("%w: payload bytes", ErrLimitExceeded)
	}
	return WireMessage{ID: id, Payload: append([]byte(nil), payload...)}, nil
}

// Decode applies an already-resolved schema to one parsed payload. Registry
// discovery is not supplied by this wrapper and is prohibited by the codec
// contract, not mechanically prevented. Decode does not bind message.ID to
// schema.Fingerprint; the caller must resolve and match that identity first.
// It copies payload but borrows target, whose partial mutation on error is codec-defined.
func (integration *CodecIntegration) Decode(
	ctx context.Context,
	schema Schema,
	message WireMessage,
	target any,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if schema.Fingerprint() == (Fingerprint{}) || message.ID.Provider == "" || message.ID.Value == "" {
		return fmt.Errorf("%w: schema and parsed provider ID are required", ErrInvalidRequest)
	}
	if len(message.Payload) > integration.limits.MaxPayloadBytes {
		return fmt.Errorf("%w: payload bytes", ErrLimitExceeded)
	}
	if err := integration.codec.Decode(ctx, schema, append([]byte(nil), message.Payload...), target); err != nil {
		return privateDiagnostic("schema registry: decode value failed", err)
	}
	return nil
}
