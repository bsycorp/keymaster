
variable "lambda_function_name" {
  description = "Lambda function name to create (override)"
  type = string
  default = ""
}

variable "resource_tags" {
  description = "Map of tags to apply to all AWS resources"
  type = map(string)
  default = {}
}

variable "artifact_s3_bucket" {
  description = "S3 bucket with existing keymaster deployment artifact (lambda zip file)"
  type = string
}

variable "artifact_s3_key" {
  description = "S3 key with existing keymaster deployment artifact (lambda zip file)"
  type = string
}

variable "allowed_invoker_arns" {
  description = "List of accounts / principals with permission to invoke km issuing api"
  type = list(string)
  default = []
}

variable "lambda_role_arn" {
  description = "Set this to override the IAM role used by the km issuing lambda"
  type = string
}

variable "configuration" {
  description = "Keymaster configuration (environment variables)"
  type = map(string)
}

variable "reserved_concurrent_executions" {
  description = "Reserved executions for each keymaster lambda"
  type = number
  default = -1
}

variable "timeout" {
  description = "Lambda timeout"
  type = number
  default = 30
}
