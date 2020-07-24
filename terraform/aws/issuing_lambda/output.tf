
output "issuing_lambda_arn" {
  value = aws_lambda_function.km.arn
  description = "The ARN of the keymaster issuing lambda."
}

