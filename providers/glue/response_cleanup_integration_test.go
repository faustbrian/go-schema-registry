//go:build integration

package glue_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsglue "github.com/aws/aws-sdk-go-v2/service/glue"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	registryglue "github.com/faustbrian/go-schema-registry/providers/glue/v2"
	schemaregistry "github.com/faustbrian/go-schema-registry/v2"
	registryavro "github.com/faustbrian/go-schema-registry/v2/formats/avro"
)

type cleanupBody struct {
	io.Reader
	closes int
}

func (body *cleanupBody) Close() error {
	body.closes++
	return nil
}

type cleanupTransport struct {
	t        *testing.T
	target   string
	requests []string
	bodies   []*cleanupBody
}

func (transport *cleanupTransport) Do(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	target := request.Header.Get("X-Amz-Target")
	transport.requests = append(transport.requests, target)
	status := http.StatusOK
	var response string
	switch target {
	case "AWSGlue.GetSchemaVersion":
		response = getVersionExchange(latestRequest(), integrationVersion1, integrationAvroV1, 1, "AVAILABLE").response
	case "AWSGlue.GetSchemaByDefinition":
		response = successByDefinition(integrationVersion1, 1)
		if transport.target == "AWSGlue.RegisterSchemaVersion" {
			status = http.StatusBadRequest
			response = smithyError("EntityNotFoundException")
		}
	case "AWSGlue.RegisterSchemaVersion":
		response = `{"SchemaVersionId":"` + integrationVersion1 + `","VersionNumber":1}`
	default:
		transport.t.Fatalf("unexpected operation %q", target)
	}
	body := &cleanupBody{Reader: strings.NewReader(response)}
	transport.bodies = append(transport.bodies, body)
	header := make(http.Header)
	header.Set("Content-Type", "application/x-amz-json-1.1")
	if status != http.StatusOK {
		header.Set("X-Amzn-Errortype", "EntityNotFoundException")
	}
	return &http.Response{StatusCode: status, Header: header, Body: body, Request: request}, nil
}

type cleanupRejection struct{}

func (*cleanupRejection) Error() string { return "private interceptor diagnostic" }

type cleanupInterceptor struct {
	target    string
	rejection *cleanupRejection
	calls     int
}

func (interceptor *cleanupInterceptor) reject(input *smithyhttp.InterceptorContext) error {
	if input.Request.Header.Get("X-Amz-Target") != interceptor.target {
		return nil
	}
	interceptor.calls++
	return interceptor.rejection
}

func (interceptor *cleanupInterceptor) AfterTransmit(_ context.Context, input *smithyhttp.InterceptorContext) error {
	return interceptor.reject(input)
}

func (interceptor *cleanupInterceptor) BeforeDeserialization(_ context.Context, input *smithyhttp.InterceptorContext) error {
	return interceptor.reject(input)
}

// Each consumed operation installs its own generated response cleanup. Exercise
// all three through the injected public client, without teardown closing bodies.
func TestProviderClosesSDKResponsesAfterInterceptorRejection(t *testing.T) {
	for _, operation := range []string{"GetSchemaVersion", "GetSchemaByDefinition", "RegisterSchemaVersion"} {
		for _, stage := range []string{"success", "AfterTransmit", "BeforeDeserialization"} {
			t.Run(operation+"/"+stage, func(t *testing.T) {
				target := "AWSGlue." + operation
				transport := &cleanupTransport{t: t, target: target}
				rejection := &cleanupRejection{}
				interceptor := &cleanupInterceptor{target: target, rejection: rejection}
				client := awsglue.NewFromConfig(aws.Config{
					Region: "eu-north-1", Credentials: aws.AnonymousCredentials{}, HTTPClient: transport,
				}, func(options *awsglue.Options) {
					options.BaseEndpoint = aws.String("https://faithful-glue.invalid")
					options.Retryer = retry.NewStandard(func(options *retry.StandardOptions) { options.MaxAttempts = 1 })
					switch stage {
					case "AfterTransmit":
						options.Interceptors.AddAfterTransmit(interceptor)
					case "BeforeDeserialization":
						options.Interceptors.AddBeforeDeserialization(interceptor)
					}
				})
				provider, err := registryglue.New(registryglue.Config{
					API: client, Scope: integrationScope, RequestTimeout: time.Second, MaxConcurrent: 1,
					Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{
						schemaregistry.FormatAvro: registryavro.New(170_000),
					},
				})
				if err != nil {
					t.Fatalf("construct provider: %v", err)
				}
				var id string
				var outcome schemaregistry.RegistrationOutcome
				if operation == "GetSchemaVersion" {
					result, resolveErr := provider.Resolve(context.Background(), schemaregistry.Latest(integrationSubject()))
					err, id = resolveErr, result.ID.Value
					if stage == "success" && (result.Version.Number != 1 || result.Lifecycle != schemaregistry.LifecycleAvailable) {
						t.Fatalf("resolved version/lifecycle = %+v/%q", result.Version, result.Lifecycle)
					}
				} else {
					result, registerErr := provider.Register(context.Background(), registerRequest(t, integrationAvroV1))
					err, id, outcome = registerErr, result.ID.Value, result.Outcome
				}
				if stage == "success" {
					if err != nil || id != integrationVersion1 {
						t.Fatalf("operation identity/error = %q/%v", id, err)
					}
					if operation == "GetSchemaByDefinition" && outcome != schemaregistry.RegistrationExisting {
						t.Fatalf("existing registration outcome = %q", outcome)
					}
					if operation == "RegisterSchemaVersion" && outcome != schemaregistry.RegistrationUnknown {
						t.Fatalf("write registration outcome = %q", outcome)
					}
				} else {
					var cause *cleanupRejection
					if !errors.Is(err, rejection) || !errors.As(err, &cause) || cause != rejection || !errors.Is(err, schemaregistry.ErrUnavailable) {
						t.Fatalf("rejection identity/category = %v", err)
					}
					if strings.Contains(err.Error(), rejection.Error()) {
						t.Fatalf("private diagnostic exposed: %v", err)
					}
					if interceptor.calls != 1 {
						t.Fatalf("target interceptor calls = %d", interceptor.calls)
					}
					if operation == "RegisterSchemaVersion" && (outcome != schemaregistry.RegistrationUnknown || !errors.Is(err, schemaregistry.ErrUnknownOutcome)) {
						t.Fatalf("ambiguous write outcome/error = %q/%v", outcome, err)
					}
				}
				expected := []string{target}
				if operation == "RegisterSchemaVersion" {
					expected = []string{"AWSGlue.GetSchemaByDefinition", target}
				}
				if len(transport.requests) != len(expected) {
					t.Fatalf("requests = %v, want %v", transport.requests, expected)
				}
				for index, want := range expected {
					if transport.requests[index] != want {
						t.Fatalf("request %d = %q, want %q", index, transport.requests[index], want)
					}
					if transport.bodies[index].closes == 0 {
						t.Errorf("response %d (%s) was not closed before return", index, want)
					}
				}
			})
		}
	}
}
