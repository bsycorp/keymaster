
locals {
  timeout              = 30
  handler              = "issuing-lambda-linux-x64"
  runtime              = "go1.x"
}

resource "aws_lambda_function" "km" {
  function_name = var.lambda_function_name
  handler       = local.handler
  runtime       = local.runtime
  s3_bucket     = var.artifact_s3_bucket
  s3_key        = var.artifact_s3_key

  role    = var.lambda_role_arn
  timeout = var.timeout

  dynamic "environment" {
    for_each = var.configuration[*]
    content {
      variables = environment.value
    }
  }

  reserved_concurrent_executions = var.reserved_concurrent_executions
  tags                           = merge({}, var.resource_tags)
}

resource "aws_lambda_permission" "allow_invoke" {
  count         = length(var.allowed_invoker_arns)
  statement_id  = "AllowClientExecution-${count.index}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.km.function_name
  principal     = var.allowed_invoker_arns[count.index]
}
