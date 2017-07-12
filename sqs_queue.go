package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/cenkalti/backoff"
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

	backoff.Retry(func() error {
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
			return err // backoff + retry
		}

		// exit backoff loop successfully
		sqsMessages = output.Messages
		return nil
	}, backoffForever())

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

	backoff.Retry(func() error {
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
			return err // backoff + retry
		}

		return nil // exit backoff loop
	}, backoffForever())
}

// backoffForever returns a BackOff that continues forever with an interval
// ranging from 1 second to 1 minute.
func backoffForever() backoff.BackOff {
	return &backoff.ExponentialBackOff{
		InitialInterval:     time.Second,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Minute,
		Clock:               backoff.SystemClock,
	}
}
