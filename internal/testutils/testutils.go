package testutils

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"

	tc "github.com/testcontainers/testcontainers-go"
)

func DockerComposeUp(t *testing.T) {
	t.Helper()

	composeFilePaths := []string{"../docker-compose.yml"}
	identifier := strings.ToLower(uuid.New().String())
	compose := tc.NewLocalDockerCompose(composeFilePaths, identifier)
	err := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke().Error
	if err != nil {
		t.Fatalf("Could not run compose file: %v - %v", composeFilePaths, err)
	}

	t.Cleanup(func() {
		err := compose.Down().Error
		if err != nil {
			t.Errorf("Could not stop services from compose file: %v - %v", composeFilePaths, err)
		}
	})
}

func LocalstackConfig(t *testing.T) aws.Config {
	t.Helper()

	const endpoint = "http://localhost:4566"
	const region = "us-east-1"

	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           endpoint,
			SigningRegion: region,
		}, nil
	})

	credentials := credentials.NewStaticCredentialsProvider("test", "test", "")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(resolver),
		config.WithCredentialsProvider(credentials),
	)
	if err != nil {
		t.Fatalf("Cannot load AWS config: %v", err)
	}

	return cfg
}

func SSMClient(t *testing.T) *ssm.Client {
	t.Helper()
	return ssm.NewFromConfig(LocalstackConfig(t))
}
