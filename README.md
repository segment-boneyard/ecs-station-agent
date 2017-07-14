# ecs-station-agent

`ecs-station-agent` reads [ECS state change events from SQS][events] and stores
both historical and up-to-date task state in Postgres. With current _and_
historical data for every ECS task that you're running or have ever run, you can
use simple SQL queries to answer many common questions:

* What tasks are currently running?
* Did a given task exit successfully yet?
* What tasks were running at a specific time in the past?
* How many tasks have failed over time?

`ecs-station-agent` is especially useful if you frequently exceed the
`DescribeTasks` API rate limit. You can also use `ecs-station-agent` to write
your own task schedulers or historical auditing tools.

Because `ecs-station-agent` is only a demo / template app, it is intentionally
missing a few things:

* Tests.

* From-scratch state reconciliation. When `ecs-station-agent` is started for the
  first time, it will not have information for tasks until they change state.

* Support for databases other than Postgres.

* Support for multi-container ECS task definitions. `ecs-station-agent` only
  records the state for the first container definition.

Feel free to extend / reuse / modify `ecs-station-agent`!

## Getting Started

To get `ecs-station-agent` up and running, you'll need three things:

### 1. CloudWatch Events > SQS pipeline

Configuring CloudWatch Events to send ECS state change events to SQS can be
a bit tricky. We've written an example Terraform configuration that does this
for you at `terraform/queue.tf`.

### 2. Postgres

You can use any ol' Postgres server, but we recommend running a dedicated RDS
instance for this. Make sure that you can connect to Postgres from your Docker
container hosts.

### 3. Docker host running `ecs-station-agent`

`ecs-station-agent` is available as a [Docker container on Docker
Hub][container]. You can run it with:

    docker run segment/ecs-station-agent -postgres <url> -sqs <url>

If you're using ECS you will likely want to create a task definition and service
for this. We created a simple Terraform template for this at
`terraform/ecs_task_definition.tf`.

## See Also

[Blox][blox] is a much more comprehensive suite of tools that can be used to
write custom ECS task schedulers. Blox requires running and maintaining an etcd
cluster in addition to many other parts. We wrote `ecs-station-agent` because we
wanted something that was minimal in terms of operation effort.

[events]: http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_cwe_events.html
[container]: https://hub.docker.com/r/segment/ecs-station-agent/
[blox]: https://blox.github.io
