package main

import (
	"flag"
	"log"
	"os"
)

var (
	// Settings
	sqsURL      = flag.String("sqs", "", "SQS queue URL containing ECS state change events")
	postgresURL = flag.String("postgres", "", "Postgres URL to store ECS task state in")

	// Globals
	queue *Queue
	store *PostgresStore
)

func init() {
	flag.Parse()

	if len(*sqsURL) == 0 || len(*postgresURL) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var err error

	queue, err = DialSQS(*sqsURL)
	if err != nil {
		log.Fatalf("Failed to dial SQS queue: %s", err.Error())
	}

	store, err = DialPostgresStore(*postgresURL)
	if err != nil {
		log.Fatalf("Failed to dial Postgres store: %s", err.Error())
	}
}

// In a perpetual loop, wait for ECS state change event messages from SQS and
// then persist them to Postgres.
func main() {
	for {
		successfullyUpdated := []Message{}

		for _, message := range queue.Receive() {
			state := message.ECSEvent.Task
			err := store.Update(state)
			if err != nil {
				log.Printf("Failed to upsert %s in Postgres: %s", state.TaskARN, err.Error())
				continue
			}
			successfullyUpdated = append(successfullyUpdated, message)
		}

		queue.Delete(successfullyUpdated)
	}
}
