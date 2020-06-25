package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bsycorp/keymaster/km/api"
	"github.com/bsycorp/keymaster/km/server"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
)

func Handler(ctx context.Context, req api.Request) (interface{}, error) {
	var km server.Server
	err := km.Configure(os.Getenv("CONFIG"))
	if err != nil {
		nerr := errors.Wrap(err,"Error loading km api configuration")
		log.Println(nerr)
		return nil, nerr
	}
	switch r := req.Payload.(type) {
	case *api.DiscoveryRequest:
		return km.HandleDiscovery(r)
	case *api.ConfigRequest:
		return km.HandleConfig(r)
	case *api.DirectSamlAuthRequest:
		return km.HandleDirectSamlAuth(r)
	case *api.DirectOidcAuthRequest:
		return km.HandleDirectOidcAuth(r)
	case *api.WorkflowStartRequest:
		return km.HandleWorkflowStart(r)
	case *api.WorkflowAuthRequest:
		return km.HandleWorkflowAuth(r)
	default:
		return nil, errors.New("unexpected request")
	}
}

func main() {
	lambda.Start(Handler)
}
