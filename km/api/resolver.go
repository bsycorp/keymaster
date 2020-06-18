package api

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func EndpointResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	if service == lambda.ServiceName && os.Getenv("KM_LAMBDA_ENDPOINT") != "" {
		return endpoints.ResolvedEndpoint{
			URL: os.Getenv("KM_LAMBDA_ENDPOINT"),
		}, nil
	}
	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}
