package commands

import (
	"context"
	"fmt"
	"github.com/bsycorp/keymaster/km/api"
	"github.com/bsycorp/keymaster/km/workflow"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"time"
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Perform keymaster authentication, in CI",
	Long: `Perform keymaster authentication, in CI.

Example:

km --issuer <issuing-lambda> --role deployment \
  --username smithb12 \
  --name "Bob Smith" \
  --email "bob.smith@awesome.com" \
  --description "enhance the magic" \
  --url "https://github.com/bsycorp/keymaster/pull/7"

All fields are required.
`,
	Run: ci,
}

var usernameFlag string
var nameFlag string
var emailFlag string
var descriptionFlag string
var detailsUrlFlag string

func init() {
	rootCmd.AddCommand(ciCmd)

	ciCmd.Flags().StringVar(&issuerFlag, "issuer", "", "target credential issuer")
	ciCmd.Flags().StringVar(&roleFlag, "role", "", "role to apply for with issuer")

	_ = ciCmd.MarkFlagRequired("issuer")
	_ = ciCmd.MarkFlagRequired("role")

	ciCmd.Flags().StringVar(&usernameFlag, "username", "", "username to associate with access request")
	ciCmd.Flags().StringVar(&nameFlag, "name", "", "human name to associate with access request")
	ciCmd.Flags().StringVar(&emailFlag, "email", "", "email address to associate with access request")
	ciCmd.Flags().StringVar(&descriptionFlag, "description", "", "describe the purpose of the access request")
	ciCmd.Flags().StringVar(&detailsUrlFlag, "url", "", "url with further details for access request")

	_ = ciCmd.MarkFlagRequired("username")
	_ = ciCmd.MarkFlagRequired("name")
	_ = ciCmd.MarkFlagRequired("email")
	_ = ciCmd.MarkFlagRequired("description")
	_ = ciCmd.MarkFlagRequired("url")
}

func ci(cmd *cobra.Command, args []string) {
	kmApi := api.NewClient(issuerFlag)
	kmApi.Debug = debugFlag

	discoveryReq := new(api.DiscoveryRequest)
	_, err := kmApi.Discovery(discoveryReq)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error calling kmApi.Discovery"))
	}

	configReq := new(api.ConfigRequest)
	configResp, err := kmApi.GetConfig(configReq)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error calling kmApi.GetConfig"))
	}

	// Now start workflow to get nonce
	kmWorkflowStartResponse, err := kmApi.WorkflowStart(&api.WorkflowStartRequest{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "error calling kmApi.WorkflowStart"))
	}
	log.Println("Started workflow with km api")

	log.Println("Target role for authentication:", roleFlag)
	targetRole := configResp.Config.FindRoleByName(roleFlag)
	if targetRole == nil {
		log.Fatalf("Target role #{roleFlag} not found in config")
	}
	workflowPolicyName := targetRole.Workflow
	configWorkflowPolicy := configResp.Config.Workflow.FindPolicyByName(workflowPolicyName)
	if configWorkflowPolicy == nil {
		log.Fatalf("workflow policy %s not found in config", workflowPolicyName)
	}
	workflowPolicy := workflow.Policy{
		Name:                configWorkflowPolicy.Name,
		IdpName:             configWorkflowPolicy.IdpName,
		RequesterCanApprove: configWorkflowPolicy.RequesterCanApprove,
		IdentifyRoles:       configWorkflowPolicy.IdentifyRoles,
		ApproverRoles:       configWorkflowPolicy.ApproverRoles,
	}

	workflowBaseUrl := configResp.Config.Workflow.BaseUrl
	log.Println("Using workflow engine:", workflowBaseUrl)
	workflowApi, err := workflow.NewClient(workflowBaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	workflowApi.Debug = debugFlag

	// And start a workflow session
	startResult, err := workflowApi.Create(context.Background(), &workflow.CreateRequest{
		IdpNonce: kmWorkflowStartResponse.IdpNonce,
		Requester: workflow.Requester{
			Name:     nameFlag,
			Username: usernameFlag,
			Email:    emailFlag,
		},
		Source: workflow.Source{
			Description: descriptionFlag,
			DetailsURI:  detailsUrlFlag,
		},
		Target: workflow.Target{
			EnvironmentName:         configResp.Config.Name,
			EnvironmentDiscoveryURI: issuerFlag,
		},
		Policy: workflowPolicy,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Now fix up the workflow URL
	log.Printf("------------------------------------------------------------------")
	log.Printf("******************************************************************")
	log.Printf("APPROVAL URL: %s", startResult.WorkflowUrl)
	log.Printf("******************************************************************")
	log.Printf("------------------------------------------------------------------")

	// Poll for assertions
	var getAssertionsResult *workflow.GetAssertionsResponse
	for {
		getAssertionsResult, err = workflowApi.GetAssertions(context.Background(), &workflow.GetAssertionsRequest{
			WorkflowId:    startResult.WorkflowId,
			WorkflowNonce: startResult.WorkflowNonce,
		})
		if err != nil {
			log.Println(errors.Wrap(err, "error calling workflowApi.GetAssertions"))
		}
		log.Printf("workflow state: %s", getAssertionsResult.Status)
		if getAssertionsResult.Status == "CREATED" {
			time.Sleep(5 * time.Second)
		} else if getAssertionsResult.Status == "COMPLETED" {
			break
		} else if getAssertionsResult.Status == "REJECTED" {
			log.Fatal("Your change request was REJECTED by a workflow approver. Exiting.")
		} else {
			log.Fatal("unexpected assertions result status:", getAssertionsResult.Status)
		}
	}
	log.Printf("got: %d assertions from workflow", len(getAssertionsResult.Assertions))

	creds, err := kmApi.WorkflowAuth(&api.WorkflowAuthRequest{
		Username:     usernameFlag,
		Role:         roleFlag,
		IdpNonce:     kmWorkflowStartResponse.IdpNonce,
		IssuingNonce: kmWorkflowStartResponse.IssuingNonce,
		Assertions:   getAssertionsResult.Assertions,
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "error calling kmApi.WorkflowAuth"))
	}

	var iamCred *api.Cred
	for _, cred := range creds.Credentials {
		if cred.Type == "iam" {
			iamCred = &cred
			break
		}
	}
	if iamCred == nil {
		log.Fatal("Got creds but no IAM cred?")
	}
	iamCredValue, ok := iamCred.Value.(*api.IAMCred)
	if !ok {
		log.Fatal("oops IAM cred is wrong type?")
	}

	awsCredsFmt := `[%s]
aws_access_key_id = %s
aws_secret_access_key = %s
aws_session_token = %s
# Keymaster issued, expires: %s
`
	exp := time.Unix(iamCred.Expiry, 0)
	localAwsCreds := fmt.Sprintf(
		awsCredsFmt,
		iamCredValue.ProfileName,
		iamCredValue.AccessKeyId,
		iamCredValue.SecretAccessKey,
		iamCredValue.SessionToken,
		exp,
	)


	existingCreds, err := ioutil.ReadFile(awsCredentialsFileFlag)
	if err != nil {
		fmt.Printf("Failed to update local credentials: %v", err)
	} else {
		log.Printf("Found existing credentials file, appending..")
		awsCredentialsIni, err := ini.Load(existingCreds, []byte(localAwsCreds))
		if err != nil {
			fmt.Printf("Failed to read existing local credentials: %v", err)
		} else {
			err = awsCredentialsIni.SaveTo(awsCredentialsFileFlag)
			if err != nil {
				fmt.Printf("Failed to update local credentials: %v", err)
			}
		}
	}
}
