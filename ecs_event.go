package main

import (
	"time"
)

// ECSEvent is an ECS state change event as sent by ECS to to SQS. It can be
// unmarshaled directly from the SQS message body.
type ECSEvent struct {
	// ID uniqely identifies this ECS event.
	ID string `json:"id"`

	// Time is when this state change event occurred.
	Time time.Time `json:"time"`

	// Task is the new state of the ECS task.
	Task ECSTaskState `json:"detail"`
}

// ECSTaskState is the state of an ECS task at a point in time.
type ECSTaskState struct {
	// TaskARN is the ARN of the task.
	TaskARN string `json:"taskArn"`

	// TaskDefinitionARN is the ARN for the task definition of this task.
	TaskDefinitionARN string `json:"taskDefinitionArn"`

	// ClusterARN is the ARN of the cluster that hosts the task.
	ClusterARN string `json:"clusterArn"`

	// ContainerInstanceARN is the ARN of the container instance that hosts the
	// task.
	ContainerInstanceARN string `json:"containerInstanceArn"`

	// CreatedAt is when the task entered the PENDING state.
	CreatedAt *time.Time `json:"createdAt"`

	// StartedAt is when the task transitioned from the PENDING to RUNNING state.
	StartedAt *time.Time `json:"startedAt"`

	// StoppedAt is when the task was stopped.
	StoppedAt *time.Time `json:"stoppedAt"`

	// StoppedReason is the reason the task stopped.
	StoppedReason string `json:"stoppedReason"`

	// DesiredStatus is the desired status of the task.
	DesiredStatus string `json:"desiredStatus"`

	// LastStatus is the last known status of the task.
	LastStatus string `json:"lastStatus"`

	// Containers is a list of the containers associated with the task.
	Containers []ECSContainer `json:"containers"`

	// Version is the version counter for the task state. It is incremented by
	// ECS every time the task state changes.
	Version int `json:"version"`
}

// ECSContainer is a Docker container that is part of an ECS task.
type ECSContainer struct {
	// ARN is the ARN of the container.
	ARN string `json:"containerArn"`

	// ExitCode is the exit code returned from the container.
	ExitCode int `json:"exitCode"`

	// LastStatus is the last known status of the container.
	LastStatus string `json:"lastStatus"`

	// Name is the name of the container.
	Name string `json:"name"`
}
