
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    sid = "LambdaServiceAssumeRole"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    effect = "Allow"
  }
}

data "aws_iam_policy_document" "km" {
  statement {
    sid = "ReadFromS3"
    actions = [
      "s3:GetObject",
    ]
    resources = var.s3_readable_objects
    effect = "Allow"
  }
  statement {
    sid = "WriteLogs"
    actions = [
      "logs:PutLogEvents",
      "logs:CreateLogStream",
      "logs:CreateLogGroup"
    ]
    resources = ["arn:aws:logs:*:*:*"]
    effect    = "Allow"
  }
  statement {
    sid = "AssumeAWSRoles"
    actions = [
      "sts:AssumeRole",
    ]
    effect    = "Allow"
    resources = var.target_role_arns
  }
}

resource "aws_iam_role" "km" {
  name               = var.role_name
  description        = "keymaster issuing lambda role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  tags               = merge({}, var.iam_principal_tags)
}

resource "aws_iam_policy" "km" {
  name        = var.role_name
  description = "keymaster iam policy"
  policy      = data.aws_iam_policy_document.km.json
}

resource "aws_iam_role_policy_attachment" "km" {
  role       = aws_iam_role.km.name
  policy_arn = aws_iam_policy.km.arn
}
