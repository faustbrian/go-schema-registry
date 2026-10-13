package confluent_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	schemaregistry "github.com/faustbrian/go-schema-registry/v3"
)

func TestDeleteFingerprintConfirmationIsSeparatePreflight(t *testing.T) {
	for _, mode := range []schemaregistry.DeletionMode{schemaregistry.DeleteSoft, schemaregistry.DeleteHard} {
		t.Run(string(mode), func(t *testing.T) {
			var gets, deletes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/subjects/subject/versions/1" {
					t.Errorf("unexpected target: %s", request.URL.Path)
					writer.WriteHeader(400)
					return
				}
				if request.Method == http.MethodGet {
					gets.Add(1)
					_, _ = writer.Write([]byte(`{"subject":"subject","version":1,"id":7,"schema":"\"string\"","schemaType":"AVRO"}`))
					return
				}
				if request.Method != http.MethodDelete {
					t.Errorf("unexpected method: %s", request.Method)
					writer.WriteHeader(400)
					return
				}
				deletes.Add(1)
				body, err := io.ReadAll(request.Body)
				if err != nil || len(body) != 0 || request.Header.Get("If-Match") != "" || gets.Load() != 2 {
					t.Error("DELETE unexpectedly carried a conditional fingerprint or bypassed preflight")
				}
				if (request.URL.Query().Get("permanent") == "true") != (mode == schemaregistry.DeleteHard) {
					t.Error("soft/hard wire mode changed")
				}
				_, _ = writer.Write([]byte(`1`))
			}))
			defer server.Close()
			provider := newTestProvider(t, server.URL)
			fingerprint := compileAvro(t).Fingerprint()
			digits := []byte(fingerprint.String())
			if digits[7] == '0' {
				digits[7] = '1'
			} else {
				digits[7] = '0'
			}
			wrong, parseErr := schemaregistry.ParseFingerprint(string(digits))
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			request := schemaregistry.DeleteRequest{Subject: schemaregistry.Subject{Name: "subject"}, Version: schemaregistry.Version{Number: 1}, Policy: schemaregistry.DeletionPolicy{Mode: mode, ExpectedFingerprint: wrong}}
			if _, err := provider.Delete(t.Context(), request); !errors.Is(err, schemaregistry.ErrConfirmationRequired) || deletes.Load() != 0 {
				t.Fatalf("mismatch sent destructive request: deletes=%d error=%v", deletes.Load(), err)
			}
			request.Policy.ExpectedFingerprint = fingerprint
			result, err := provider.Delete(t.Context(), request)
			want := schemaregistry.LifecycleDeleting
			if mode == schemaregistry.DeleteHard {
				want = schemaregistry.LifecycleDeleted
			}
			if err != nil || result.Lifecycle != want || deletes.Load() != 1 {
				t.Fatalf("matching preflight operation: result=%+v deletes=%d error=%v", result, deletes.Load(), err)
			}
		})
	}
}
