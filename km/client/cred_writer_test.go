package client

import (
	"github.com/bsycorp/keymaster/km/api"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var c1 = api.Cred{
	Name:   "nonprod-deployment",
	Type:   "iam",
	Expiry: 1,
	Value: &api.IAMCred{
		ProfileName:     "Foo",
		RoleArn:         "Bar",
		RoleSessionName: "123",
		AccessKeyId:     "abc",
		SecretAccessKey: "def",
		SessionToken:    "ghi",
	},
}
var c2 = api.Cred{
	Name:   "nonprod-ro",
	Type:   "iam",
	Expiry: 1,
	Value: &api.IAMCred{
		ProfileName:     "FooX",
		RoleArn:         "BarX",
		RoleSessionName: "123X",
		AccessKeyId:     "abcX",
		SecretAccessKey: "defX",
		SessionToken:    "ghiX",
	},
}

func TestSaveIAMCredentialsSimple(t *testing.T) {
	// No existing credentials file, no set profile name
	credsFile := "scratch/foo"
	opts1 := &CredWriterOptions{
		AwsSetProfileName:  "",
		AwsCredentialsFile: credsFile,
	}
	err := SaveIAMCredentials(opts1, []api.Cred{c1})
	assert.Nil(t, err)
	expect1 := `[Foo]
aws_access_key_id     = abc
aws_secret_access_key = def
aws_session_token     = ghi

`
	fooData, err := ioutil.ReadFile(credsFile)
	assert.Nil(t, err)
	assert.Equal(t, expect1, string(fooData))
	assert.NoError(t, os.Remove(credsFile))
}

func TestSaveIAMCredentialsWithForcedProfileName(t *testing.T) {
	// No existing credentials file, force profile name
	credsFile := "scratch/bar"
	opts1 := &CredWriterOptions{
		AwsSetProfileName:  "default",
		AwsCredentialsFile: credsFile,
	}
	err := SaveIAMCredentials(opts1, []api.Cred{c1})
	assert.Nil(t, err)
	expect1 := `[default]
aws_access_key_id     = abc
aws_secret_access_key = def
aws_session_token     = ghi

`
	fooData, err := ioutil.ReadFile(credsFile)
	assert.Nil(t, err)
	assert.Equal(t, expect1, string(fooData))
	assert.NoError(t, os.Remove(credsFile))
}

func TestSaveIAMCredentialsMultiple(t *testing.T) {
	// No existing credentials file, set multiple creds, no set profile name
	credsFile := "scratch/baz"
	opts1 := &CredWriterOptions{
		AwsSetProfileName:  "",
		AwsCredentialsFile: credsFile,
	}
	err := SaveIAMCredentials(opts1, []api.Cred{c1, c2})
	assert.Nil(t, err)
	expect1 := `[Foo]
aws_access_key_id     = abc
aws_secret_access_key = def
aws_session_token     = ghi

[FooX]
aws_access_key_id     = abcX
aws_secret_access_key = defX
aws_session_token     = ghiX

`
	fooData, err := ioutil.ReadFile(credsFile)
	assert.Nil(t, err)
	assert.Equal(t, expect1, string(fooData))
	assert.NoError(t, os.Remove(credsFile))
}
