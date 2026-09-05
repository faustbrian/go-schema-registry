package glue_test

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsglue "github.com/aws/aws-sdk-go-v2/service/glue"
	schemaregistry "github.com/faustbrian/go-schema-registry"
	"github.com/faustbrian/go-schema-registry/formats/avro"
	registryglue "github.com/faustbrian/go-schema-registry/providers/glue"
)

func ExampleNew() {
	client := awsglue.NewFromConfig(aws.Config{Region: "eu-north-1"})
	provider, err := registryglue.New(registryglue.Config{
		API:            client,
		Scope:          "eu-north-1:123456789012:events",
		RequestTimeout: 5 * time.Second,
		MaxConcurrent:  8,
		Canonicalizers: map[schemaregistry.Format]schemaregistry.Canonicalizer{
			schemaregistry.FormatAvro: avro.New(170_000),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(provider.Capabilities().Provider)
	// Output: aws-glue
}
