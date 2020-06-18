package api

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

type Client struct {
	//    * Function name - my-function (name-only), my-function:v1 (with alias).
	//    * Function ARN - arn:aws:lambda:us-west-2:123456789012:function:my-function.
	//    * Partial ARN - 123456789012:function:my-function.
	FunctionName string
	lambdaClient *lambda.Lambda
	Debug        int
}

func NewClient(target string) *Client {
	c := new(Client)
	_, disableSSL := os.LookupEnv("KM_DISABLE_SSL")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			EndpointResolver: endpoints.ResolverFunc(EndpointResolver),
			DisableSSL:       &disableSSL,
		},
		SharedConfigState: session.SharedConfigEnable,
	}))
	c.FunctionName = target
	c.lambdaClient = lambda.New(sess) // TODO: region? Or can that come from env?
	return c
}

func (c *Client) Discovery(req *DiscoveryRequest) (*DiscoveryResponse, error) {
	resp := new(DiscoveryResponse)
	err := c.rpc(&Request{Type: "discovery", Payload: req}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetConfig(req *ConfigRequest) (*ConfigResponse, error) {
	resp := new(ConfigResponse)
	err := c.rpc(&Request{Type: "config", Payload: req}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DirectSamlAuth(req *DirectSamlAuthRequest) (*DirectAuthResponse, error) {
	return nil, nil
}

func (c *Client) WorkflowStart(req *WorkflowStartRequest) (*WorkflowStartResponse, error) {
	resp := new(WorkflowStartResponse)
	err := c.rpc(&Request{Type: "workflow_start", Payload: req}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) WorkflowAuth(req *WorkflowAuthRequest) (*WorkflowAuthResponse, error) {
	resp := new(WorkflowAuthResponse)
	err := c.rpc(&Request{Type: "workflow_auth", Payload: req}, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) isError(resp *lambda.InvokeOutput) error {
	if resp.FunctionError != nil {
		return errors.Errorf("function error: %s: response payload: %s",
			*resp.FunctionError, string(resp.Payload))
	} else if *resp.StatusCode != 200 {
		return errors.Errorf("bad status code: %d, response payload: %s",
			*resp.StatusCode, string(resp.Payload))
	}
	return nil
}

func (c *Client) rpc(req interface{}, resp interface{}) error {
	if c.Debug > 0 {
		log.Println("rpc request: ", spew.Sdump(req))
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "rpc marshal")
	}
	result, err := c.lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(c.FunctionName),
		Payload:      payload,
	})
	if err != nil {
		return errors.Wrap(err, "rpc invoke")
	}
	if err = c.isError(result); err != nil {
		return errors.Wrap(err, "rpc error")
	}
	err = json.Unmarshal(result.Payload, resp)
	if err != nil {
		if c.Debug > 0 {
			log.Println("rpc raw response:" + string(result.Payload))
		}
		return errors.Wrap(err, "rpc unmarshal")
	}
	if c.Debug > 0 {
		log.Println("rpc response:", spew.Sdump(resp))
	}
	return nil
}
