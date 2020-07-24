variable "role_name" {
  type = string
  description = "Name to use for role + policy"
}

variable "iam_principal_tags" {
  description = "Map of tags to apply to Keymaster IAM role"
  type = map(string)
  default = {}
}

variable "target_role_arns" {
  description = "List of IAM roles which km may assume"
  type = list(string)
  default = []
}

variable "s3_readable_objects" {
  description = "List of S3 resources (objects/buckets) which issuing lambda may read"
  type = list(string)
  default = []
}
