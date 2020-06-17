package api

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lambda"
	"os"
)

func EndpointResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	if service == lambda.ServiceID && os.Getenv("LAMBDA_ENDPOINT") != "" {
		return endpoints.ResolvedEndpoint{
			URL:           os.Getenv("LAMBDA_ENDPOINT"),
		}, nil
	}
	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}
