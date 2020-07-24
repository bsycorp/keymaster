variable "ip_oracle_sign_zip" {
	default = "../../../build/ip-oracle-sign.zip"
}

variable "ip_oracle_key_arn" {
  default = "*"
}

variable "resource_name_prefix" {
	default = "kmv2"
}

variable "lambda_sign_handler" {
	default = "ip-oracle-sign-linux-x64"
}

variable "timeout" {
	description = "Lambda timeout"
	type = number
	default = 30
}
