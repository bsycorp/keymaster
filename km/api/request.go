package api

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type Request struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type DiscoveryRequest struct {}

type DiscoveryResponse struct {}

type ConfigRequest struct {
}

type ConfigResponse struct {
	Version string       `json:"version"`
	Config  ConfigPublic `json:"config"`
}

type DirectSamlAuthRequest struct {
	RequestedRole string  `json:"requested_role"`
	SAMLResponse  string  `json:"saml_response"`
	SigAlg        string  `json:"sig_alg"`
	Signature     string  `json:"signature"`
	RelayState    *string `json:"relay_state,omitempty"`
}

type DirectOidcAuthRequest struct {
}

type DirectAuthResponse struct {
	Credentials map[string][]byte `json:"result"`
}

type WorkflowStartRequest struct {
}

type WorkflowStartResponse struct {
	IssuingNonce string `json:"issuing_nonce"`
	IdpNonce string `json:"idp_nonce"`
}

type WorkflowAuthRequest struct {
	Username string `json:"username"`
	Role string `json:"role"`
	IssuingNonce string `json:"issuing_nonce"`
	IdpNonce string `json:"idp_nonce"`
	Assertions []string `json:"assertions"`
}

type WorkflowAuthResponse struct {
	Credentials []Cred `json:"credentials"`
}

func (c *Request) UnmarshalJSON(data []byte) error {
	var t struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	c.Type = t.Type
	var payload interface{}
	switch c.Type {
	case "discovery":
		payload = &DiscoveryRequest{}
	case "config":
		payload = &ConfigRequest{}
	case "direct_saml_auth":
		payload = &DirectSamlAuthRequest{}
	case "direct_oidc_auth":
		payload = &DirectOidcAuthRequest{}
	case "workflow_start":
		payload = &WorkflowStartRequest{}
	case "workflow_auth":
		payload = &WorkflowAuthRequest{}
	default:
		return errors.New("unknown operation type: " + c.Type)
	}
	err = json.Unmarshal(t.Payload, payload)
	if err != nil {
		return err
	}
	c.Payload = payload
	return nil
}
