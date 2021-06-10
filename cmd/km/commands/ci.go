package commands

import (
	"context"
	"github.com/bsycorp/keymaster/km/api"
	"github.com/bsycorp/keymaster/km/client"
	"github.com/bsycorp/keymaster/km/workflow"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

const GetAssertionsPollDelay 5 * time.Second

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

	// Run workflow to get assertions.
	assertions := runWorkflow(targetRole, &configResp.Config, kmWorkflowStartResponse.IdpNonce)

	creds, err := kmApi.WorkflowAuth(&api.WorkflowAuthRequest{
		Username:     usernameFlag,
		Role:         roleFlag,
		IdpNonce:     kmWorkflowStartResponse.IdpNonce,
		IssuingNonce: kmWorkflowStartResponse.IssuingNonce,
		Assertions:   assertions,
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "error calling kmApi.WorkflowAuth"))
	}

	credWriterOptions := client.CredWriterOptions{
		AwsSetProfileName: awsSetProfileNameFlag,
		AwsCredentialsFile: awsCredentialsFileFlag,
	}
	err = client.SaveIAMCredentials(&credWriterOptions, creds.Credentials)
	if err != nil {
		log.Errorf( "error writing IAM credentials file: %v", err)
	}
}

func runWorkflow(targetRole *api.RoleConfig, config *api.ConfigPublic, idpNonce string) []string {
	workflowPolicyName := targetRole.Workflow
	configWorkflowPolicy := config.Workflow.FindPolicyByName(workflowPolicyName)
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

	// If no approval or identify required, skip all workflow. This may also
	// be useful in emergencies if workflow is down...
	if len(workflowPolicy.IdentifyRoles) == 0 && len(workflowPolicy.ApproverRoles) == 0 {
		log.Println("Skipping workflow - no identify or approval required")
		return []string{}
	}

	workflowBaseUrl := config.Workflow.BaseUrl
	log.Println("Using workflow engine:", workflowBaseUrl)
	workflowApi, err := workflow.NewClient(workflowBaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	workflowApi.Debug = debugFlag

	// And start a workflow session
	startResult, err := workflowApi.Create(context.Background(), &workflow.CreateRequest{
		IdpNonce: idpNonce,
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
			EnvironmentName:         config.Name,
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
			time.Sleep(GetAssertionsPollDelay)
			continue
		}
		log.Printf("workflow state: %s", getAssertionsResult.Status)
		if getAssertionsResult.Status == "CREATED" {
			time.Sleep(GetAssertionsPollDelay)
		} else if getAssertionsResult.Status == "COMPLETED" {
			break
		} else if getAssertionsResult.Status == "REJECTED" {
			log.Fatal("Your change request was REJECTED by a workflow approver. Exiting.")
		} else {
			log.Fatal("unexpected assertions result status:", getAssertionsResult.Status)
		}
	}
	log.Printf("got: %d assertions from workflow", len(getAssertionsResult.Assertions))
	return getAssertionsResult.Assertions
}