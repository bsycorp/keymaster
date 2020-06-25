package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL         *url.URL
	HttpClient      *http.Client
	PolicyEncrypter string // TODO: key, signingmethod etc
	Debug           int
}

func NewClient(baseUrl string) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing workflow client URL")
	}
	return &Client{
		BaseURL:         u,
		HttpClient:      http.DefaultClient,
		PolicyEncrypter: "", // TODO
	}, nil
}

type Requester struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Source struct {
	// Source Key? For gitlab?
	Description string `json:"description"`
	DetailsURI  string `json:"details_uri"`
}

type Target struct {
	EnvironmentName         string `json:"environment_name"`
	EnvironmentDiscoveryURI string `json:"environment_discovery_uri"`
}

type Policy struct {
	Name                string         `json:"name"`
	IdpName             string         `json:"idp_name"`
	RequesterCanApprove bool           `json:"requester_can_approve"`
	IdentifyRoles       map[string]int `json:"identify_roles"`
	ApproverRoles       map[string]int `json:"approver_roles"`
}

type CreateRequest struct {
	IdpNonce  string    `json:"idp_nonce"`
	Requester Requester `json:"requester"`
	Source    Source    `json:"source"`
	Target    Target    `json:"target"`
	// To be encrypted with workflow public key
	// TODO: change type here, separation of concerns...
	Policy Policy `json:"policy"`
}

type CreateResponse struct {
	WorkflowId    string `json:"workflow_id"`
	WorkflowUrl   string `json:"workflow_url"`
	WorkflowNonce string `json:"workflow_nonce"`
}

type GetDetailsRequest struct {
	// Workflow id
}

type GetDetailsResponse struct {
	// State: pending / completed
}

type GetAssertionsRequest struct {
	WorkflowId    string `json:"workflow_id"`
	WorkflowNonce string `json:"workflow_nonce"`
}

type GetAssertionsResponse struct {
	// Bag of SAML assertions. Could be wrapped(?)
	//Workflow Workflow `json:"workflow"` //???
	Status     string   `json:"status"`     //???
	Assertions []string `json:"assertions"` // Resulting IDP assertions
}

func (c *Client) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	var resp CreateResponse
	err := c.call(ctx, "POST", "/1/workflow/create", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "workflow create error")
	}
	return &resp, err
}

func (c *Client) GetDetails(ctx context.Context, req *GetDetailsRequest) (*GetDetailsResponse, error) {
	var resp GetDetailsResponse
	err := c.call(ctx, "POST", "/1/workflow/getDetails", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "workflow getdetails error")
	}
	return &resp, err
}

func (c *Client) GetAssertions(ctx context.Context, req *GetAssertionsRequest) (*GetAssertionsResponse, error) {
	var resp GetAssertionsResponse
	err := c.call(ctx, "POST", "/1/workflow/getAssertions", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "workflow getassertions error")
	}
	return &resp, err
}

func (c *Client) call(ctx context.Context, method string, path string, req interface{}, resp interface{}) error {
	if c.Debug > 0 {
		log.Printf("workflow request: %s: %s", method, path)
		log.Println("workflow request:", spew.Sdump(req))
	}
	httpReq, err := c.newRequest(ctx, method, path, req)
	if err != nil {
		return errors.Wrap(err, "http request construction error")
	}
	_, err = c.do(httpReq, resp)
	if err != nil {
		return errors.Wrap(err, "http error")
	}
	if c.Debug > 0 {
		log.Println("workflow response:", spew.Sdump(resp))
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading workflow response body")
	}
	if c.Debug > 1 {
		log.Printf("raw status code: %d", resp.StatusCode)
		log.Printf("raw response: %s", string(body))
	}
	if resp.StatusCode >= 400 { // http 4xx, 5xx
		return nil, errors.Errorf("server error: StatusCode: %d: Body: %s",
			resp.StatusCode, string(body))
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal error")
	}

	return resp, nil
}
