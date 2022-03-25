package sqs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	"github.com/pkg/errors"
)

type Provider struct {
	client *sqs.Client
}

func NewProvider(client *sqs.Client) *Provider {
	return &Provider{client: client}
}

func (p *Provider) SendMessage(ctx context.Context, msg models.Message, queue string) (string, error) {
	// TEMP :: queue URL

	out, err := p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(msg.Data),
	})
	if err != nil {
		return "", errors.Wrapf(err, "unable to send message to %v", queue)
	}

	return aws.ToString(out.MessageId), nil
}

func (p *Provider) PollForNewMessages(ctx context.Context, queue string) ([]*models.Message, error) {
	out, err := p.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queue),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to receive messages from queue %v", queue)
	}

	if len(out.Messages) == 0 {
		return nil, nil
	}

	messagesToReturn := make([]*models.Message, 0, len(out.Messages))
	messagesToDelete := make([]types.DeleteMessageBatchRequestEntry, 0, len(out.Messages))
	for _, msg := range out.Messages {
		newLocalMessage := &models.Message{
			Queue:    queue,
			ExtID:    aws.ToString(msg.MessageId),
			Received: time.Now(),
			Data:     aws.ToString(msg.Body),
		}
		messagesToReturn = append(messagesToReturn, newLocalMessage)

		// Pull the message from the queue
		// TODO: should this be determined by the caller?
		messagesToDelete = append(messagesToDelete, types.DeleteMessageBatchRequestEntry{
			Id:            msg.MessageId,
			ReceiptHandle: msg.ReceiptHandle,
		})
	}

	if _, err := p.client.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(queue),
		Entries:  messagesToDelete,
	}); err != nil {
		log.Printf("error deleting messages from queue: %v", err)
	}

	return messagesToReturn, nil
}
