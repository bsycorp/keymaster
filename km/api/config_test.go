package api

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestLoadSampleConfigs(t *testing.T) {
	expected := Config{
		Name:    "fooproject_nonprod",
		Version: "1.0",
		Idp: []IdpConfig{
			{
				Name: "nonprod",
				Type: "saml",
				Config: &IdpConfigSaml{
					Certificate: "-----BEGIN CERTIFICATE-----\nMIICnTCCAYUCBgFfA+Q72DANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDDAdjbHVz\ndGVyMB4XDTE3MTAxMDAxMjUxMFoXDTI3MTAxMDAxMjY1MFowEjEQMA4GA1UEAwwH\nY2x1c3RlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMdVYTd1h7fa\nu6/uCgboFyFdoRSWFEHP0Iq9GUWA69g2x+QDqZikSv/JqPwtJBAm+dxdXfOd0RKT\n4ypK09PUNy542kJ+Qwgzwif0ZIEKTYOVS8VvwzZv6BjzwDzSBS/LmdcK8WgRGwgh\n62QgjIYQdGd+wrYN0tOQb6EzINWMs1bq9bFjeFegDG94p/MZ1YWRVXF6h/euq/ym\ngJQc7yvUn5cy6l47tT1ARrCzpUF8Ss4eVhNlLDaz5WSzZ4P1Q+bPe4Iax//zMr/J\n62aqmcf/YuVKIINLa5ML+QFW2B+mR0xky8jwWJiwU5gJzDzLoiNQZ3TJxcfvQaT1\nPuC8ksM9bd0CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAvnrKy75SHGEAIPORf2QC\nNxqWi6Qc/Pl1gHSGHd9nPcIn7u2dRmoq45XWAr55yVZqT/FWshOII504YuFJCQF5\nfyOGKy00jVmaOEIPqyLRA0wf4AsZk607Y2CVZIl1JGwuYx5rHgZ2kf1M4Qxvnhl/\nOUkMrW+VosBgIrqiKWd53Y5TnHaX/q+hYoa/GmRXq0JTJOX+5C11YX9G4rsI7o3c\nMP19yto+e+d5myXu3POAvx4VG07LlWWk3cow2xuiw4zJbZVmK6KO2rMk66WJpfQu\nEmyLmLPjKTmhoskvaHhvSoW6h06Uth3Lf6UHHsAkdzeU+mw0g2Zb2dPlDqz4IV4t\ncg==\n-----END CERTIFICATE-----\n",
					Audience:    "keymaster-saml",
					UsernameAttr: "name",
					EmailAttr:    "name",
					GroupsAttr:   "groups",
					RedirectURI:  "https://workflow.int.btr.place/1/saml/approve",
				},
			},
		},
		Roles: []RoleConfig{
			{
				Name:            "cloudengineer",
				Credentials:     []string{"ssh-all", "kube", "aws-admin"},
				Workflow:        "cloudengineer",
				ValidForSeconds: 7200,
			},
			{
				Name:            "developer",
				Credentials:     []string{"ssh-jumpbox", "kube", "aws-ro"},
				Workflow:        "developer",
				ValidForSeconds: 7200,
			},
			{
				Name:            "deployment",
				Credentials:     []string{"kube", "aws-admin"},
				Workflow:        "deploy_with_approval",
				ValidForSeconds: 3600,
				CredentialDelivery: RoleCredentialDeliveryConfig{
					KmsWrapWith: "arn:aws:kms:ap-southeast-2:062921715532:key/95a6a059-8281-4280-8500-caf8cc217367",
				},
			},
		},
		Credentials: []CredentialsConfig{
			{
				Name: "ssh-jumpbox",
				Type: "ssh_ca",
				Config: &CredentialsConfigSSH{
					CAKey:      "s3://my-bucket/sshca.key",
					Principals: []string{"$idpuser"},
				},
			},
			{
				Name: "ssh-all",
				Type: "ssh_ca",
				Config: &CredentialsConfigSSH{
					CAKey:      "s3://my-bucket/sshca.key",
					Principals: []string{"$idpuser", "core", "ec2-user"},
				},
			},
			{
				Name: "kube-user",
				Type: "kubernetes",
				Config: &CredentialsConfigKube{
					CAKey: "s3://my-bucket/kubeca.key",
				},
			},
			{
				Name: "kube-admin",
				Type: "kubernetes",
				Config: &CredentialsConfigKube{
					CAKey: "s3://my-bucket/kubeca.key",
				},
			},
			{
				Name: "aws-ro",
				Type: "iam_assume_role",
				Config: &CredentialsConfigIAMAssumeRole{
					TargetRole: "arn:aws:iam::062921715666:role/ReadOnly",
				},
			},
			{
				Name: "aws-admin",
				Type: "iam_assume_role",
				Config: &CredentialsConfigIAMAssumeRole{
					TargetRole: "Administrator",
				},
			},
		},
		Workflow: WorkflowConfig{
			BaseUrl: "https://workflow.int.btr.place/",
			Policies: []WorkflowPolicyConfig{
				{
					Name:                "deploy_with_identify",
					IdpName:             "nonprod",
					RequesterCanApprove: false,
					IdentifyRoles: map[string]int{
						"adfs_role_deployer": 1,
					},
				},
				{
					Name:                "deploy_with_approval",
					IdpName:             "nonprod",
					RequesterCanApprove: false,
					ApproverRoles: map[string]int{
						"adfs_role_approver": 1,
					},
				},
				{
					Name:                "deploy_with_identify_and_approval",
					IdpName:             "nonprod",
					RequesterCanApprove: false,
					IdentifyRoles: map[string]int{
						"adfs_role_deployer": 1,
					},
					ApproverRoles: map[string]int{
						"adfs_role_approver": 1,
					},
				},
				{
					Name:    "developer",
					IdpName: "nonprod",
					IdentifyRoles: map[string]int{
						"adfs_role_developer": 1,
					},
				},
				{
					Name:    "cloudengineer",
					IdpName: "nonprod",
					IdentifyRoles: map[string]int{
						"adfs_role_cloudengineer": 1,
					},
				},
			},
		},
		AccessControl: AccessControlConfig{
			IPOracle: IPOracleConfig{
				WhiteListCidrs: []string{"192.168.0.0/24", "172.16.0.0/12", "10.0.0.0/8"},
			},
		},
	}
	data, err := ioutil.ReadFile("./testdata/example_api_config.yaml")
	assert.NoError(t, err)
	var result Config
	err = yaml.Unmarshal([]byte(data), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestConfig_FindCredentialByName(t *testing.T) {
	config := Config{
		Credentials: []CredentialsConfig{
			{
				Name: "ssh-all",
				Type: "ssh_ca",
				Config: &CredentialsConfigSSH{
					CAKey:      "s3://my-bucket/sshca.key",
					Principals: []string{"$idpuser", "core", "ec2-user"},
				},
			},
		},
	}
	credConfig := config.FindCredentialByName("ssh-all")
	assert.NotNil(t, credConfig)

	assert.Nil(t, config.FindCredentialByName("does-not-exist"))
}

func TestConfig_FindRoleByName(t *testing.T) {
	config := Config{
		Roles: []RoleConfig{
			{
				Name:            "developer",
				Credentials:     []string{"ssh-jumpbox", "kube", "aws-ro"},
				Workflow:        "developer",
				ValidForSeconds: 7200,
			},
		},
	}
	credConfig := config.FindRoleByName("developer")
	assert.NotNil(t, credConfig)

	assert.Nil(t, config.FindCredentialByName("does-not-exist"))
}

func TestIdpConfig_UnmarshalJSON(t *testing.T) {
	testCases := map[string]IdpConfig{
		"t1": {
			Type: "saml",
			Name: "my-idp",
			Config: &IdpConfigSaml{
				Certificate: "-----BEGIN CERTIFICATE-----\nMIICnTCCAYUCBgFfA+Q72DANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDDAdjbHVz\ndGVyMB4XDTE3MTAxMDAxMjUxMFoXDTI3MTAxMDAxMjY1MFowEjEQMA4GA1UEAwwH\nY2x1c3RlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMdVYTd1h7fa\nu6/uCgboFyFdoRSWFEHP0Iq9GUWA69g2x+QDqZikSv/JqPwtJBAm+dxdXfOd0RKT\n4ypK09PUNy542kJ+Qwgzwif0ZIEKTYOVS8VvwzZv6BjzwDzSBS/LmdcK8WgRGwgh\n62QgjIYQdGd+wrYN0tOQb6EzINWMs1bq9bFjeFegDG94p/MZ1YWRVXF6h/euq/ym\ngJQc7yvUn5cy6l47tT1ARrCzpUF8Ss4eVhNlLDaz5WSzZ4P1Q+bPe4Iax//zMr/J\n62aqmcf/YuVKIINLa5ML+QFW2B+mR0xky8jwWJiwU5gJzDzLoiNQZ3TJxcfvQaT1\nPuC8ksM9bd0CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAvnrKy75SHGEAIPORf2QC\nNxqWi6Qc/Pl1gHSGHd9nPcIn7u2dRmoq45XWAr55yVZqT/FWshOII504YuFJCQF5\nfyOGKy00jVmaOEIPqyLRA0wf4AsZk607Y2CVZIl1JGwuYx5rHgZ2kf1M4Qxvnhl/\nOUkMrW+VosBgIrqiKWd53Y5TnHaX/q+hYoa/GmRXq0JTJOX+5C11YX9G4rsI7o3c\nMP19yto+e+d5myXu3POAvx4VG07LlWWk3cow2xuiw4zJbZVmK6KO2rMk66WJpfQu\nEmyLmLPjKTmhoskvaHhvSoW6h06Uth3Lf6UHHsAkdzeU+mw0g2Zb2dPlDqz4IV4t\ncg==\n-----END CERTIFICATE-----\n",
				Audience:    "keymaster-saml",
				UsernameAttr: "name",
				EmailAttr:    "name",
				GroupsAttr:   "groups",
				RedirectURI:  "https://workflow.int.btr.place/1/saml/approve",
			},
		},
	}

	// Unmarshal c -> c2, check c == c2
	for _, c := range testCases {
		b, err := json.Marshal(c)
		assert.NoError(t, err)
		assert.NotEmpty(t, b)

		var c2 IdpConfig
		err = json.Unmarshal(b, &c2)
		assert.NoError(t, err)

		assert.Equal(t, c, c2)
	}
}

func TestCredentialsConfig_UnmarshalJSON(t *testing.T) {
	testCases := map[string]CredentialsConfig{
		"ssh1": {
			Name: "ssh-example",
			Type: "ssh_ca",
			Config: &CredentialsConfigSSH{
				CAKey: "my-ssh-ca-key",
			},
		},
		"kube1": {
			Name:   "kube-example",
			Type:   "kubernetes",
			Config: &CredentialsConfigKube{},
		},
		"iam_assumerole1": {
			Name:   "iam-assumerole-example",
			Type:   "iam_assume_role",
			Config: &CredentialsConfigIAMAssumeRole{},
		},
		"iam_user1": {
			Name:   "iam-user-example",
			Type:   "iam_user",
			Config: &CredentialsConfigIAMUser{},
		},
	}

	// Unmarshal c -> c2, check c == c2
	for _, c := range testCases {
		b, err := json.Marshal(c)
		assert.NoError(t, err)
		assert.NotEmpty(t, b)

		var c2 CredentialsConfig
		err = json.Unmarshal(b, &c2)
		assert.NoError(t, err)

		assert.Equal(t, c, c2)
	}
}

func TestCertificateLoading(t *testing.T) {
	// Test that certificates get loaded via util.Load(), which should
	// resolve file://, data:// and s3:// uris
	c := Config{
		Name:    "fooproject_nonprod",
		Version: "1.0",
		Idp: []IdpConfig{
			{
				Name: "nonprod",
				Type: "saml",
				Config: &IdpConfigSaml{
					Certificate: "data://Zm9v",
				},
			},
		},
	}
	err := c.NormaliseAndLoad()
	assert.Nil(t, err)
	assert.Equal(t, "foo", c.Idp[0].Config.(*IdpConfigSaml).Certificate)
}
