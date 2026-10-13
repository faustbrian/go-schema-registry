package confluent_test

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/faustbrian/go-schema-registry/providers/confluent/v3"
)

const frameMachineIntBits = strconv.IntSize

// Build wire bytes independently of FrameMessage, which only accepts native ints.
func signed64Frame(values ...int64) []byte {
	frame := []byte{0, 0, 0, 0, 1}
	for _, value := range values {
		encoded := uint64(value) << 1
		if value < 0 {
			encoded = uint64(-(value+1))*2 + 1
		}
		frame = binary.AppendUvarint(frame, encoded)
	}
	return frame
}

func TestProtobufFrameMachineIntegerAdmission(t *testing.T) {
	framer, err := confluent.NewProtobufFramer("machine-int", 32, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, length := range []int64{-(1 << 32), (1 << 32) + 1, math.MaxInt64} {
		t.Run("count/"+strconv.FormatInt(length, 10), func(t *testing.T) {
			_, _, _, err := framer.UnframeMessage(t.Context(), signed64Frame(length, 0))
			if !errors.Is(err, confluent.ErrInvalidFrame) {
				t.Fatalf("invalid signed64 count survived native conversion: %v", err)
			}
		})
	}
	for _, value := range []int64{-(1 << 32), -1, 0, math.MaxInt32, math.MaxInt32 + 1, (1 << 32) + 7, math.MaxInt64} {
		t.Run("index/"+strconv.FormatInt(value, 10), func(t *testing.T) {
			_, indexes, _, err := framer.UnframeMessage(t.Context(), signed64Frame(1, value))
			admitted := value >= 0 && (frameMachineIntBits == 64 || value <= math.MaxInt32)
			if !admitted {
				if !errors.Is(err, confluent.ErrInvalidFrame) {
					t.Fatalf("out-of-domain index survived native conversion: indexes=%v error=%v", indexes, err)
				}
				return
			}
			if err != nil || len(indexes) != 1 || int64(indexes[0]) != value {
				t.Fatalf("admitted index changed: indexes=%v error=%v", indexes, err)
			}
		})
	}
}
