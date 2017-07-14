# -- 1. Create the CloudWatch Event rule

resource "aws_cloudwatch_event_rule" "ecs_task_state_events" {
  name        = "ecs-task-state-change"
  description = "Send all ECS task state changes"

  event_pattern = <<PATTERN
    {
      "source": [
        "aws.ecs"
      ],
      "detail-type": [
        "ECS Task State Change"
      ]
    }
PATTERN
}

# -- 2. Create the SQS queue

resource "aws_sqs_queue" "ecs_task_state_events" {
  name                       = "ecs-task-state-events"
  visibility_timeout_seconds = 10
}

# -- 3. Allow CloudWatch Events to send messages to SQS

resource "aws_sqs_queue_policy" "ecs_task_state_events_cloudwatch" {
  queue_url = "${aws_sqs_queue.ecs_task_state_events.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "${aws_sqs_queue.ecs_task_state_events.arn}/SQSDefaultPolicy",
  "Statement": [
    {
      "Sid": "TrustCWEToSendEventsToMyQueue",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "${aws_sqs_queue.ecs_task_state_events.arn}",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "${aws_cloudwatch_event_rule.ecs_task_state_events.arn}"
        }
      }
    }
  ]
}
POLICY
}

# -- 4. Send the CloudWatch events to SQS

resource "aws_cloudwatch_event_target" "ecs_task_state_events" {
  rule = "${aws_cloudwatch_event_rule.ecs_task_state_events.name}"
  arn  = "${aws_sqs_queue.ecs_task_state_events.arn}"
}

# -- Outputs

output "sqs_url" {
  url = "${aws_sqs_queue.ecs_task_state_events.id}"
}
