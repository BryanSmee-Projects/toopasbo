
resource "aws_iam_role" "lambda_role" {
  name = "lambda_exec_role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [{
      Action    = "sts:AssumeRole",
      Effect    = "Allow",
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_policy" "lambda_policy" {
  name        = "lambda_policy"
  description = "Policy for Lambda function"

  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action   = "secretsmanager:GetSecretValue",
        Effect   = "Allow",
        Resource = data.aws_secretsmanager_secret.toobo_secrets.arn
      },
      {
        Action   = ["dynamodb:PutItem", "dynamodb:GetItem", "dynamodb:UpdateItem", "dynamodb:DeleteItem", "dynamodb:Scan", "dynamodb:Query"],
        Effect   = "Allow",
        Resource = aws_dynamodb_table.chat_config.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_policy.arn
}

resource "aws_dynamodb_table" "chat_config" {
  name           = "toobo-chat-config"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"
  attribute {
    name = "id"
    type = "S"
  }
}

data "aws_secretsmanager_secret" "toobo_secrets" {
  name = "toobo-secrets"
}

resource "aws_iam_policy_attachment" "lambda_policy" {
  name       = "toobo_lambda_policy_attachment"
  roles      = [aws_iam_role.lambda_role.name]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "archive_file" "webhook_zip" {
  type        = "zip"
  source_dir  = "../dist/webhook"
  output_path = "webhook-lambda.zip"
}

resource "aws_lambda_function" "toobo" {
  function_name = "toobo-webhook"
  role          = aws_iam_role.lambda_role.arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = data.archive_file.webhook_zip.output_path
  timeout       = 300
  source_code_hash = data.archive_file.webhook_zip.output_base64sha256

  environment {
    variables = {
      RUNTIME_ENVIRONMENT = "aws"
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.chat_config.name
      AWS_SECRET_NAME     = data.aws_secretsmanager_secret.toobo_secrets.name
    }
  }
}

data "archive_file" "daily_zip" {
  type        = "zip"
  source_dir  = "../dist/daily"
  output_path = "daily-lambda.zip"
}

resource "aws_lambda_function" "daily-toobo" {
  function_name = "toobo-daily"
  role          = aws_iam_role.lambda_role.arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = data.archive_file.daily_zip.output_path
  timeout       = 300
  source_code_hash = data.archive_file.daily_zip.output_base64sha256

  environment {
    variables = {
      RUNTIME_ENVIRONMENT = "aws"
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.chat_config.name
      AWS_SECRET_NAME     = data.aws_secretsmanager_secret.toobo_secrets.name
    }
  }
}

data "archive_file" "weekly_zip" {
  type        = "zip"
  source_dir  = "../dist/weekly"
  output_path = "weekly-lambda.zip"
}

resource "aws_lambda_function" "weekly-toobo" {
  function_name = "toobo-weekly"
  role          = aws_iam_role.lambda_role.arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = data.archive_file.weekly_zip.output_path
  timeout       = 600
  source_code_hash = data.archive_file.weekly_zip.output_base64sha256

  environment {
    variables = {
      RUNTIME_ENVIRONMENT = "aws"
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.chat_config.name
      AWS_SECRET_NAME     = data.aws_secretsmanager_secret.toobo_secrets.name
    }
  }
}

resource "aws_api_gateway_rest_api" "lambda_api" {
  name        = "TooBoAPI"
  description = "API Gateway for TooBo"
}

resource "aws_api_gateway_resource" "api_resource" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api.id
  parent_id   = aws_api_gateway_rest_api.lambda_api.root_resource_id
  path_part   = "webhook"
}

resource "aws_api_gateway_method" "api_method" {
  rest_api_id   = aws_api_gateway_rest_api.lambda_api.id
  resource_id   = aws_api_gateway_resource.api_resource.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_lambda_permission" "api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.toobo.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.lambda_api.execution_arn}/*/*"
}

resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.lambda_api.id
  resource_id             = aws_api_gateway_resource.api_resource.id
  http_method             = aws_api_gateway_method.api_method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.toobo.invoke_arn
}

resource "aws_api_gateway_deployment" "api_deployment" {
  depends_on  = [aws_api_gateway_integration.lambda_integration]
  rest_api_id = aws_api_gateway_rest_api.lambda_api.id
  stage_name  = "prod"
}

resource "aws_cloudwatch_event_rule" "daily" {
  name        = "every_day_rule"
  description = "trigger everyday at 6:00"

  schedule_expression = "cron(0 6 * * ? *)"
}

resource "aws_cloudwatch_event_target" "lambda_target" {
  rule      = aws_cloudwatch_event_rule.daily.name
  target_id = "SendToLambda"
  arn       = aws_lambda_function.daily-toobo.arn
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.daily-toobo.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.daily.arn
}

resource "aws_cloudwatch_event_rule" "weekly" {
  name        = "every_week_rule"
  description = "trigger every week at 6:00 on Monday"

  schedule_expression = "cron(0 6 ? * MON *)"
}

resource "aws_cloudwatch_event_target" "weekly_target" {
  rule      = aws_cloudwatch_event_rule.weekly.name
  target_id = "SendToLambda"
  arn       = aws_lambda_function.weekly-toobo.arn
}

resource "aws_lambda_permission" "allow_weekly_eventbridge" {
  statement_id  = "AllowExecutionFromWeeklyEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.weekly-toobo.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.weekly.arn
}
