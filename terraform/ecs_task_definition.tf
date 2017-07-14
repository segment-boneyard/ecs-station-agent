resource "aws_ecs_task_definition" "ecs_station_agent" {
  family = "ecs-station-agent"

  container_definitions = <<EOF
    [
      {
        "name": "ecs-station-agent",
        "image": "segment/ecs-station-agent",
        "cpu": 256,
        "memory": 128,
        "command": ["--sqs", "<SQS_URL>", "--postgres", "<POSTGRES_URL>"],
      }
    ]
    EOF
}

resource "aws_ecs_service" "ecs_station_agent" {
  name            = "ecs-station-agent"
  task_definition = "${aws_ecs_task_definition.ecs_station_agent.arn}"
  desired_count   = 1
}
