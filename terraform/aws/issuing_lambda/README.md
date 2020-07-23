Usage:

```hcl
module "issuing_lambda" {
  source = "github.com/bsycorp/keymaster//terraform/aws/issuing_lambda?ref=v0.1.3"

  # The environment label will be added to all named resources
  env_label   = "myproject-npe"

  # Keymaster configuration file
  configuration = {
     CONFIG: "s3://km-myproject-npe/km-myproject-npe.yaml"
  }

  # List of target roles that the lambda may issue creds for
  target_role_arns = [
   "arn:aws:iam::218296299700:role/test_env_admin"
  ]

  # List of client accounts that may invoke issuing lambda
  client_account_arns = [
   "arn:aws:iam::062921715666:root",   # myproj-dev-01
  ]

  # Enable creation of the configuration bucket and upload
  # of the configuration file
  config_bucket_enable = true
  config_file_upload_enable = true
  config_file_name = "${path.module}/km-myproject-npe.yaml"

  resource_tags = {
    Env          = "myproject-npe"
    Created-By   = "yourteam@you.com"
  }
}
```

Where `km-myproject-npe.yaml` contains e.g:

```
name: nonprod
version: "1.0"
idp:
  - name: adfs-local
    type: saml
    config:
      audience: keymaster-saml
      username_attr: name
      email_attr: name     # ignored
      groups_attr: groups
      redirect_uri: https://your.workflow.engine/1/saml/approve
      # Cert may be specified as s3:// file:// or raw data
      certificate: |
        -----BEGIN CERTIFICATE-----
        MIICnTCCAYUCBgFfA+Q72DANBgkqhkiG9w0BAQsFADASMRAwDgYDVQQDDAdjbHVz
        dGVyMB4XDTE3MTAxMDAxMjUxMFoXDTI3MTAxMDAxMjY1MFowEjEQMA4GA1UEAwwH
        Y2x1c3RlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMdVYTd1h7fa
        u6/uCgboFyFdoRSWFEHP0Iq9GUWA69g2x+QDqZikSv/JqPwtJBAm+dxdXfOd0RKT
        4ypK09PUNy542kJ+Qwgzwif0ZIEKTYOVS8VvwzZv6BjzwDzSBS/LmdcK8WgRGwgh
        62QgjIYQdGd+wrYN0tOQb6EzINWMs1bq9bFjeFegDG94p/MZ1YWRVXF6h/euq/ym
        gJQc7yvUn5cy6l47tT1ARrCzpUF8Ss4eVhNlLDaz5WSzZ4P1Q+bPe4Iax//zMr/J
        62aqmcf/YuVKIINLa5ML+QFW2B+mR0xky8jwWJiwU5gJzDzLoiNQZ3TJxcfvQaT1
        PuC8ksM9bd0CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAvnrKy75SHGEAIPORf2QC
        NxqWi6Qc/Pl1gHSGHd9nPcIn7u2dRmoq45XWAr55yVZqT/FWshOII504YuFJCQF5
        fyOGKy00jVmaOEIPqyLRA0wf4AsZk607Y2CVZIl1JGwuYx5rHgZ2kf1M4Qxvnhl/
        OUkMrW+VosBgIrqiKWd53Y5TnHaX/q+hYoa/GmRXq0JTJOX+5C11YX9G4rsI7o3c
        MP19yto+e+d5myXu3POAvx4VG07LlWWk3cow2xuiw4zJbZVmK6KO2rMk66WJpfQu
        EmyLmLPjKTmhoskvaHhvSoW6h06Uth3Lf6UHHsAkdzeU+mw0g2Zb2dPlDqz4IV4t
        cg==
        -----END CERTIFICATE-----
roles:
  - name: deployment
    credentials: [aws-admin]
    workflow: deploy_with_approval
    valid_for_seconds: 3600
workflow:
  base_url: https://your.workflow.engine/
  policies:
    - name: deploy_with_approval
      requester_can_approve: false
      approver_roles:
        Approvers: 1
credentials:
  - name: aws-admin
    type: iam_assume_role
    config:
      # Can be role ARN or role name, if only name is given the
      # role will be looked up in the target account.
      target_role: arn:aws:iam::218296299766:role/test_env_admin

```

## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| artifact\_file | Local path to lambda deployment package. Conflicts with artifact\_s3\* | `string` | `null` | no |
| artifact\_s3\_bucket | S3 bucket with existing keymaster deployment artifact (lambda zip file) | `string` | `null` | no |
| artifact\_s3\_key | S3 key with existing keymaster deployment artifcat (lambda zip file) | `string` | `null` | no |
| client\_account\_arns | List of accounts with permission to invoke km issuing api | `list(string)` | `[]` | no |
| config\_bucket\_enable | Create the config bucket | `bool` | `false` | no |
| config\_bucket\_name | Name of bucket to store configuration file | `string` | `""` | no |
| config\_file\_name | Name of local file to upload for km configuration | `string` | `""` | no |
| config\_file\_upload\_enable | Enable uploading a configuration file for km | `bool` | `false` | no |
| configuration | Keymaster configuration (environment variables) | `map(string)` | n/a | yes |
| env\_label | The tag label of the environment km will be deployed into (e.g. btr-place) | `string` | n/a | yes |
| lambda\_function\_name | Lambda function name to create | `string` | `""` | no |
| lambda\_role\_arn | Set this to override the IAM role used by the km issuing lambda | `string` | `""` | no |
| reserved\_concurrent\_executions | Reserved executions for each keymaster lambda | `number` | `-1` | no |
| resource\_tags | Map of tags to apply to all AWS resources | `map(string)` | `{}` | no |
| target\_role\_arns | List of roles which km may issue credentials for | `list(string)` | `[]` | no |
| timeout | Lambda timeout | `number` | `30` | no |

## Outputs

| Name | Description |
|------|-------------|
| configuration\_bucket\_name | The name of the km configuration bucket. Will be empty if not configured. |
| issuing\_lambda\_arn | The ARN of the keymaster issuing lambda. |

