//go:build integration

package glue_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsglue "github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/smithy-go/tracing"
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
	throttle bool
	signed   bool
}

func (transport *cleanupTransport) Do(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	target := request.Header.Get("X-Amz-Target")
	transport.requests = append(transport.requests, target)
	if transport.signed && !strings.HasPrefix(request.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") {
		transport.t.Fatal("configured credentials did not sign request")
	}
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
	if transport.throttle && len(transport.requests) == 1 {
		status = http.StatusBadRequest
		response = `{"__type":"ThrottlingException","Message":"throttled"}`
	}
	body := &cleanupBody{Reader: strings.NewReader(response)}
	transport.bodies = append(transport.bodies, body)
	header := make(http.Header)
	header.Set("Content-Type", "application/x-amz-json-1.1")
	if status != http.StatusOK {
		errorType := "EntityNotFoundException"
		if transport.throttle && len(transport.requests) == 1 {
			errorType = "ThrottlingException"
		}
		header.Set("X-Amzn-Errortype", errorType)
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
				var spans []string
				rejection := &cleanupRejection{}
				interceptor := &cleanupInterceptor{target: target, rejection: rejection}
				client := awsglue.NewFromConfig(aws.Config{
					Region: "eu-north-1", Credentials: aws.AnonymousCredentials{}, HTTPClient: transport,
				}, func(options *awsglue.Options) {
					options.BaseEndpoint = aws.String("https://faithful-glue.invalid")
					options.TracerProvider = cleanupTraceProvider{names: &spans}
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
				if countSpans(spans, "RetryLoop") != len(expected) {
					t.Errorf("RetryLoop spans = %d, want %d; spans = %v", countSpans(spans, "RetryLoop"), len(expected), spans)
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

type cleanupTraceProvider struct{ names *[]string }

func (p cleanupTraceProvider) Tracer(scope string, options ...tracing.TracerOption) tracing.Tracer {
	return cleanupTracer{base: tracing.NopTracerProvider{}.Tracer(scope, options...), names: p.names}
}

type cleanupTracer struct {
	base  tracing.Tracer
	names *[]string
}

func (t cleanupTracer) StartSpan(ctx context.Context, name string, options ...tracing.SpanOption) (context.Context, tracing.Span) {
	*t.names = append(*t.names, name)
	return t.base.StartSpan(ctx, name, options...)
}
func countSpans(names []string, want string) int {
	count := 0
	for _, name := range names {
		if name == want {
			count++
		}
	}
	return count
}
func cleanupProvider(t *testing.T, client *awsglue.Client) *registryglue.Provider {
	t.Helper()
	provider, err := registryglue.New(registryglue.Config{
		API: client, Scope: integrationScope, RequestTimeout: time.Second, MaxConcurrent: 1,
		Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{schemaregistry.FormatAvro: registryavro.New(170_000)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return provider
}

func TestProviderKeepsOneSDKRetryLoopAcrossAttempts(t *testing.T) {
	transport := &cleanupTransport{t: t, throttle: true}
	var spans []string
	client := awsglue.New(awsglue.Options{
		Region: "eu-north-1", Credentials: aws.AnonymousCredentials{}, HTTPClient: transport,
		BaseEndpoint: aws.String("https://faithful-glue.invalid"), TracerProvider: cleanupTraceProvider{names: &spans},
		Retryer: retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = 2
			o.Backoff = retry.BackoffDelayerFunc(func(int, error) (time.Duration, error) { return 0, nil })
		}),
	})
	result, err := cleanupProvider(t, client).Resolve(context.Background(), schemaregistry.Latest(integrationSubject()))
	if err != nil || result.ID.Value != integrationVersion1 {
		t.Fatalf("resolve after throttle = %q/%v", result.ID.Value, err)
	}
	if countSpans(spans, "RetryLoop") != 1 || countSpans(spans, "Attempt") != 2 || len(transport.requests) != 2 {
		t.Fatalf("retry spans/requests = %v/%v", spans, transport.requests)
	}
	for _, body := range transport.bodies {
		if body.closes == 0 {
			t.Error("retry response was not closed before return")
		}
	}
}

// Default-chain composition is verified offline; no profile, metadata service,
// or account credentials from the surrounding environment may participate.
func TestProviderWithDefaultConfigCredentials(t *testing.T) {
	for _, kind := range []string{"environment", "shared-file"} {
		t.Run(kind, func(t *testing.T) {
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if strings.HasPrefix(key, "AWS_") {
					t.Setenv(key, "")
				}
			}
			directory := t.TempDir()
			configFile, credentialsFile := filepath.Join(directory, "config"), filepath.Join(directory, "credentials")
			if err := os.WriteFile(configFile, nil, 0600); err != nil {
				t.Fatal(err)
			}
			profile := ""
			if kind == "shared-file" {
				profile = "[fixture]\naws_access_key_id=fixture-access\naws_secret_access_key=fixture-secret\n"
				t.Setenv("AWS_PROFILE", "fixture")
			} else {
				t.Setenv("AWS_ACCESS_KEY_ID", "fixture-access")
				t.Setenv("AWS_SECRET_ACCESS_KEY", "fixture-secret")
			}
			if err := os.WriteFile(credentialsFile, []byte(profile), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("AWS_CONFIG_FILE", configFile)
			t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile)
			t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			transport := &cleanupTransport{t: t, signed: true}
			configuration, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("eu-north-1"), awsconfig.WithHTTPClient(transport))
			if err != nil {
				t.Fatalf("load fixture configuration: %v", err)
			}
			credentials, err := configuration.Credentials.Retrieve(ctx)
			if err != nil || credentials.AccessKeyID != "fixture-access" || credentials.SecretAccessKey != "fixture-secret" {
				t.Fatal("fixture credential resolution failed")
			}
			client := awsglue.NewFromConfig(configuration, func(o *awsglue.Options) {
				o.BaseEndpoint = aws.String("https://faithful-glue.invalid")
				o.RetryMaxAttempts = 1
			})
			result, err := cleanupProvider(t, client).Resolve(ctx, schemaregistry.Latest(integrationSubject()))
			if err != nil || result.ID.Value != integrationVersion1 {
				t.Fatalf("configured resolve = %q/%v", result.ID.Value, err)
			}
			if len(transport.requests) != 1 || len(transport.bodies) != 1 || transport.bodies[0].closes == 0 {
				t.Fatal("configured request/cleanup contract failed")
			}
		})
	}
}
