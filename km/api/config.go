package api

import (
	"encoding/json"
	"github.com/bsycorp/keymaster/km/util"
	"github.com/pkg/errors"
)

type Config struct {
	Name          string              `json:"name"`
	Version       string              `json:"version"`
	Idp           []IdpConfig         `json:"idp"`
	Roles         []RoleConfig        `json:"roles"`
	Workflow      WorkflowConfig      `json:"workflow"`
	Credentials   []CredentialsConfig `json:"credentials"`
	AccessControl AccessControlConfig `json:"access_control"`
}

func (c *Config) NormaliseAndLoad() error {
	// We default to version 1.0 if not specified
	if c.Version == "" {
		c.Version = "1.0"
	}

	// If there's no IDP name specified in a policy we will
	// just use the first IDP.
	policies := c.Workflow.Policies
	if len(c.Idp) > 0 {
		for i := range policies {
			if policies[i].IdpName == "" {
				policies[i].IdpName = c.Idp[0].Name
			}
		}
	}

	// SAML certificates may be indirect, load them if they are
	for _, idpConfig := range c.Idp {
		if samlIdp, ok := idpConfig.Config.(*IdpConfigSaml); ok {
			certData, err := util.Load(samlIdp.Certificate)
			if err != nil {
				return err
			}
			samlIdp.Certificate = string(certData)
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Version != "1.0" {
		return errors.Errorf("unsupported version: %s", c.Version)
	}
	if len(c.Idp) > 1 {
		// TODO: multiple IDP support
		return errors.New("only 1 IDP is supported")
	}
	return nil
}

func (c *Config) FindCredentialByName(name string) *CredentialsConfig {
	for _, i := range c.Credentials {
		if i.Name == name {
			return &i
		}
	}
	return nil
}

func (c *Config) FindRoleByName(name string) *RoleConfig {
	for _, i := range c.Roles {
		if i.Name == name {
			return &i
		}
	}
	return nil
}

type ConfigPublic struct {
	Name     string         `json:"name"`
	Idp      []IdpConfig    `json:"idp"`
	Roles    []RoleConfig   `json:"roles"`
	Workflow WorkflowConfig `json:"workflow"`
}

type IdpConfig struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type IdpConfigSaml struct {
	Certificate  string `json:"certificate"`
	Audience     string `json:"audience"`
	UsernameAttr string `json:"username_attr"`
	EmailAttr    string `json:"email_attr"`
	GroupsAttr   string `json:"groups_attr"`
	RedirectURI  string `json:"redirect_uri"`
}

type IdpConfigOidc struct {
	// Not implemented
}

type RoleConfig struct {
	Name               string                       `json:"name"`
	Workflow           string                       `json:"workflow"`
	Credentials        []string                     `json:"credentials"`
	ValidForSeconds    int                          `json:"valid_for_seconds"`
	CredentialDelivery RoleCredentialDeliveryConfig `json:"credential_delivery"`
	ClientDefaults     string                       `json:"client_defaults"`
}

func (c *ConfigPublic) FindRoleByName(name string) *RoleConfig {
	for _, p := range c.Roles {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

type RoleCredentialDeliveryConfig struct {
	KmsWrapWith string `json:"kms_wrap_with"`
}

type CredentialsConfig struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type CredentialsConfigSSH struct {
	CAKey      string   `json:"ca_key"`
	Principals []string `json:"principals"`
}

type CredentialsConfigKube struct {
	CAKey string `json:"ca_key"`
}

type CredentialsConfigIAMAssumeRole struct {
	TargetRole string `json:"target_role"`
}

type CredentialsConfigIAMUser struct {
	// TODO: IAM user support
}

type WorkflowConfig struct {
	BaseUrl  string                 `json:"base_url"`
	Policies []WorkflowPolicyConfig `json:"policies"`
}

func (wc *WorkflowConfig) FindPolicyByName(name string) *WorkflowPolicyConfig {
	for _, p := range wc.Policies {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

type WorkflowPolicyConfig struct {
	Name                string         `json:"name"`
	IdpName             string         `json:"idp_name"`
	RequesterCanApprove bool           `json:"requester_can_approve"`
	IdentifyRoles       map[string]int `json:"identify_roles"`
	ApproverRoles       map[string]int `json:"approver_roles"`
}

type AccessControlConfig struct {
	IPOracle IPOracleConfig `json:"ip_oracle"`
}

type IPOracleConfig struct {
	WhiteListCidrs []string `json:"whitelist_cidrs"`
}

func (c *IdpConfig) UnmarshalJSON(data []byte) error {
	var t struct {
		Name          string          `json:"name"`
		Type          string          `json:"type"`
		UntypedConfig json.RawMessage `json:"config"`
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	c.Name = t.Name
	c.Type = t.Type
	var config interface{}
	switch c.Type {
	case "saml":
		config = &IdpConfigSaml{}
	case "oidc":
		config = &IdpConfigOidc{}
	default:
		return errors.New("unknown idp type: " + c.Type)
	}
	err = json.Unmarshal(t.UntypedConfig, config)
	if err != nil {
		return err
	}
	c.Config = config
	return nil
}

func (c *CredentialsConfig) UnmarshalJSON(data []byte) error {
	var t struct {
		Name          string          `json:"name"`
		Type          string          `json:"type"`
		UntypedConfig json.RawMessage `json:"config"`
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	c.Name = t.Name
	c.Type = t.Type
	var config interface{}
	switch c.Type {
	case "ssh_ca":
		config = &CredentialsConfigSSH{}
	case "kubernetes":
		config = &CredentialsConfigKube{}
	case "iam_assume_role":
		config = &CredentialsConfigIAMAssumeRole{}
	case "iam_user":
		config = &CredentialsConfigIAMUser{}
	default:
		return errors.New("unknown credential type: " + c.Type)
	}
	err = json.Unmarshal(t.UntypedConfig, config)
	if err != nil {
		return err
	}
	c.Config = config
	return nil
}
