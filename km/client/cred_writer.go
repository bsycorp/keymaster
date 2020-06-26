package client

import (
	"fmt"
	"github.com/bsycorp/keymaster/km/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
)

type CredWriterOptions struct {
	AwsSetProfileName string
	AwsCredentialsFile string
}

func SaveIAMCredentials(options *CredWriterOptions, creds []api.Cred) error {
	// Pluck IAM Creds
	var iamCreds []*api.IAMCred
	for _, cred := range creds {
		if cred.Type == "iam" {
			iamCred, ok := cred.Value.(*api.IAMCred)
			if !ok {
				log.Errorf("failed to cast credential to IAM credential!")
			} else {
				iamCreds = append(iamCreds, iamCred)
			}
		}
	}

	// Handle the "set profile name" flag
	if options.AwsSetProfileName != "" {
		if len(iamCreds) == 0 {
			log.Warnf("no iam creds to write for profile: %v", options.AwsSetProfileName)
			return nil
		} else {
			if len(iamCreds) > 1 {
				log.Warnf("got too many iam creds; expected 1, got: %v", len(iamCreds))
			}
			tmp := *iamCreds[0]
			log.Printf("renaming iam credential %v -> %v", tmp.ProfileName, options.AwsSetProfileName)
			tmp.ProfileName = options.AwsSetProfileName
			iamCreds = []*api.IAMCred{&tmp}
		}
	}

	// Format credentials
	credEntries := make([]string, len(iamCreds))
	for _, c := range iamCreds {
		log.Printf("creating iam credential: %v", c.ProfileName)
		credEntries = append(credEntries, formatIAMCred(c))
	}

	return writeIAMCredentialsFile(options.AwsCredentialsFile, credEntries)
}

func formatIAMCred(iamCred *api.IAMCred) string {
	awsCredsFmt := `[%s]
aws_access_key_id = %s
aws_secret_access_key = %s
aws_session_token = %s
`
	strAwsCreds := fmt.Sprintf(
		awsCredsFmt,
		iamCred.ProfileName,
		iamCred.AccessKeyId,
		iamCred.SecretAccessKey,
		iamCred.SessionToken,
	)
	return strAwsCreds
}

func writeIAMCredentialsFile(credsFile string, credsToAdd []string) error {
	extraCreds := make([]interface{}, len(credsToAdd))
	for i := range credsToAdd {
		extraCreds[i] = []byte(credsToAdd[i])
	}
	existingCreds, err := ioutil.ReadFile(credsFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("no existing credentials file: %v", credsFile)
		} else {
			return errors.Wrap(err, "failed to open AWS credentials file")
		}
	}
	existingCreds = []byte{}
	awsCredentialsIni, err := ini.Load(existingCreds, extraCreds...)
	if err != nil {
		return errors.Wrap(err, "failed to load existing AWS credentials")
	} else {
		err = awsCredentialsIni.SaveTo(credsFile)
		if err != nil {
			return errors.Wrap(err, "failed to update AWS credentials file")
		}
	}
	return nil
}
