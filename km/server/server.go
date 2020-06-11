package server

import (
	"encoding/json"
	"github.com/bsycorp/keymaster/km/api"
	"github.com/bsycorp/keymaster/km/creds"
	"github.com/bsycorp/keymaster/km/idp/saml"
	"github.com/bsycorp/keymaster/km/util"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Server struct {
	Config api.Config
}

func (s *Server) Configure(config string) error {
	var err error
	var tmpConfig api.Config

	// Load config, maybe "by reference" (config env var might be a
	// literal or a reference to a bucket or file).
	// Then we load as YAML or JSON.
	configData, err := util.Load(config)
	if err != nil {
		return err
	}
	if strings.HasPrefix(string(configData), "{") {
		err = json.Unmarshal(configData, &tmpConfig)
		if err != nil {
			return err
		}
	} else {
		err = yaml.Unmarshal(configData, &tmpConfig)
		if err != nil {
			return err
		}
	}

	tmpConfig.Normalise()
	err = tmpConfig.Validate()
	if err != nil {
		return err
	}
	s.Config = tmpConfig
	return nil
}

func (s *Server) HandleDiscovery(req *api.DiscoveryRequest) (*api.DiscoveryResponse, error) {
	var resp api.DiscoveryResponse
	return &resp, nil
}

func (s *Server) HandleConfig(req *api.ConfigRequest) (*api.ConfigResponse, error) {
	// Copy the public parts of our configuration.
	var resp api.ConfigResponse
	resp.Config = api.ConfigPublic{
		Name:     s.Config.Name,
		Idp:      s.Config.Idp,
		Roles:    s.Config.Roles,
		Workflow: s.Config.Workflow,
	}
	return &resp, nil
}

func (s *Server) HandleDirectSamlAuth(eq *api.DirectSamlAuthRequest) (*api.DirectAuthResponse, error) {
	return nil, errors.New("Not implemented")
}

func (s *Server) HandleDirectOidcAuth(req *api.DirectOidcAuthRequest) (*api.DirectAuthResponse, error) {
	return nil, errors.New("Not implemented")
}

func (s *Server) HandleWorkflowStart(req *api.WorkflowStartRequest) (*api.WorkflowStartResponse, error) {
	// TODO: this will be a JWT in future
	uu := uuid.New()
	uu2 := uuid.New()
	return &api.WorkflowStartResponse{
		IssuingNonce: uu.String(),
		IdpNonce:     "x" + uu2.String(),
	}, nil
}

func (s *Server) HandleWorkflowAuth(req *api.WorkflowAuthRequest) (*api.WorkflowAuthResponse, error) {
	// Find the requested role
	role := s.Config.FindRoleByName(req.Role)
	if role == nil {
		return nil, errors.Errorf("requested role not found: %s", req.Role)
	}
	// Find the workflow policy for the requested role
	rolePolicy := s.Config.Workflow.FindPolicyByName(role.Workflow)
	if rolePolicy == nil {
		return nil, errors.Errorf("requested role policy not found: %s", role.Workflow)
	}
	// Validate that there are no identify roles
	if len(rolePolicy.IdentifyRoles) > 0 {
		return nil, errors.New("requested role requires identification; not supported")
	}
	// Validate that there is just one approval role
	if len(rolePolicy.ApproverRoles) > 1 {
		return nil, errors.New("multiple approver support not implemented")
	}
	// There should be as many IDP assertions as approvers
	if len(req.Assertions) != len(rolePolicy.ApproverRoles) {
		return nil, errors.New("wrong number of saml assertions submitted")
	}
	// Ensure there is just 1 IDP in configuration
	if len(s.Config.Idp) > 1 {
		return nil, errors.New("multiple IDP support not implemented")
	}

	// TODO: verify issuing nonce
	// TODO: verify idp nonce

	idpConfig := s.Config.Idp[0]
	idpSamlConfig := idpConfig.Config.(*api.IdpConfigSaml)
	sp := &saml.AssertionProcessor{
		CAData:       []byte(idpSamlConfig.Certificate),
		Audience:     idpSamlConfig.Audience,
		UsernameAttr: idpSamlConfig.UsernameAttr,
		EmailAttr:    idpSamlConfig.EmailAttr,
		GroupsAttr:   idpSamlConfig.GroupsAttr,
		RedirectURI:  idpSamlConfig.RedirectURI,
		DisableNameIDValidation: true,
	}
	err := sp.Init()
	if err != nil {
		return nil, errors.Wrap(err, "saml init error")
	}
	userInfos, err := sp.Process(req.IdpNonce, req.Assertions)
	if err != nil {
		return nil, errors.Wrap(err, "saml validation error")
	}

	// Count approvals from IDP assertions
	approvals := make(map[string]int)
	for _, userInfo := range userInfos {
		log.Println("Processing assertion from:", userInfo)
		approvalsFromUser := 0
		for _, groupName := range userInfo.Groups {
			_, found := rolePolicy.ApproverRoles[groupName]
			if found {
				approvals[groupName]++
				approvalsFromUser++
			}
		}
		// One assertion should represent just 1 approval from relevant group
		if approvalsFromUser == 0 {
			i := 0
			requiredGroups := make([]string, len(rolePolicy.ApproverRoles))
			for k := range rolePolicy.ApproverRoles {
				requiredGroups[i] = k
				i++
			}
			return nil, errors.Errorf("assertion with no valid approval groups from: %s got: %s want: %s",
				userInfo.Username, userInfo.Groups, requiredGroups)
		}
		if approvalsFromUser > 1 {
			return nil, errors.Errorf("assertion meets more than 1 approval group from: %s", userInfo.Username)
		}
	}
	// Validate that the required number of approvals were met
	for groupName, requiredApprovals := range rolePolicy.ApproverRoles {
		actualApprovals := approvals[groupName]
		if actualApprovals < requiredApprovals {
			return nil, errors.Errorf("not enough approvals, want: %d, got: %d",
				requiredApprovals, actualApprovals)
		}
	}

	userInfo := api.AuthInfo{
		Environment: s.Config.Name,
		Role:        req.Role,
		Username:    req.Username,
		ValidFor:    role.ValidForSeconds,
	}
	credIssuer, err := creds.NewFromConfig(role, &s.Config)
	if err != nil {
		return nil, errors.Wrap(err, "during issuer configuration")
	}
	issuedCreds, err := credIssuer.IssueFor(&userInfo)
	if err != nil {
		return nil, errors.Wrap(err, "during issuance")
	}
	return &api.WorkflowAuthResponse{
		Credentials: issuedCreds,
	}, nil
}
