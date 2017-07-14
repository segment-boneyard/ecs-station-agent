package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

const (
	// sqsMaxRetries is the maximum number of retries that the go-aws-sdk will
	// attempt when requests fail due to retryable errors (timeouts, etc.)
	sqsMaxRetries = 10

	// sqsMaxNumberOfMessages is the maximum number of SQS messages that will be
	// received by one ReceiveMessage request. The SQS-enforced maximum is 10.
	sqsMaxNumberOfMessages = 10

	// sqsWaitTimeSeconds is the amount of time that our ReceiveMessage request
	// will wait for messages if there are none immediately available. If this
	// number of seconds pass without any messages being received, the
	// ReceiveMessage request will return successfully with an empty message
	// array. The SQS-enforced maximum is 20 seconds.
	//
	// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-long-polling.html
	sqsWaitTimeSeconds = 20
)

// DialSQS returns an SQS-backed queue of ECS task state changes.
func DialSQS(queueURL string) (*SQSQueue, error) {
	sess, err := session.NewSession(&aws.Config{
		MaxRetries: aws.Int(sqsMaxRetries),
	})
	if err != nil {
		return nil, err
	}

	return &SQSQueue{
		queueURL: queueURL,
		service:  sqs.New(sess),
	}, nil
}

// SQSQueue is a simplified interface to an SQS queue. All methods use an
// indefinite backoff to retry perpetually when errors are encountered instead
// of returning errors.
type SQSQueue struct {
	queueURL string
	service  sqsiface.SQSAPI
}

// SQSMessage is a single received SQS message and its ECSEvent body.
type SQSMessage struct {
	ECSEvent      ECSEvent
	id            string
	receiptHandle string
}

// Receive returns 0 or more successfully "received" SQS messages. It never
// returns an error and will instead retry perpetually until an error is not
// returned by the SQS API.
func (q *SQSQueue) Receive() (messages []SQSMessage) {
	var sqsMessages []*sqs.Message

	tryForever(func() error {
		output, err := q.service.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(q.queueURL),
			MaxNumberOfMessages: aws.Int64(sqsMaxNumberOfMessages),
			WaitTimeSeconds:     aws.Int64(sqsWaitTimeSeconds),
			AttributeNames: []*string{
				aws.String("ApproximateReceiveCount"),
				aws.String("SentTimestamp"),
			},
		})
		if err != nil {
			log.Printf("ReceiveMessage failed: %s", err.Error())
			return err // retry
		}

		sqsMessages = output.Messages
		return nil // exit retry loop
	})

	return parseBodies(sqsMessages)
}

func parseBodies(sqsMessages []*sqs.Message) (messages []SQSMessage) {
	for _, sqsMessage := range sqsMessages {
		message := SQSMessage{
			id:            *sqsMessage.MessageId,
			receiptHandle: *sqsMessage.ReceiptHandle,
		}

		err := json.Unmarshal([]byte(*sqsMessage.Body), &message.ECSEvent)
		if err != nil {
			log.Printf("failed to unmarshal ECSEvent from SQS message body: %s", err.Error())
			continue
		}

		messages = append(messages, message)
	}

	return messages
}

// Delete takes 0 or more SQS messages and deletes them from SQS. It never
// returns an error and will instead retry perpetually until an error is not
// returned.
func (q *SQSQueue) Delete(messages []SQSMessage) {
	if len(messages) == 0 {
		return
	}

	tryForever(func() error {
		input := &sqs.DeleteMessageBatchInput{}

		for _, message := range messages {
			input.Entries = append(input.Entries, &sqs.DeleteMessageBatchRequestEntry{
				Id:            aws.String(message.id),
				ReceiptHandle: aws.String(message.receiptHandle),
			})
		}

		_, err := q.service.DeleteMessageBatch(input)
		if err != nil {
			log.Printf("DeleteMessageBatch failed: %s", err.Error())
			return err // retry
		}

		return nil // exit retry loop
	})
}

const (
	retryMin = 100 * time.Millisecond
	retryMax = 5 * time.Second
)

// tryForever calls the given function until it doesn't return an error. It
// uses a basic exponential backoff algorithm.
func tryForever(fn func() error) {
	sleep := retryMin
	for {
		if fn() == nil {
			return
		}

		time.Sleep(sleep)
		sleep = 2 * sleep
		if sleep > retryMax {
			sleep = retryMax
		}
	}
}
