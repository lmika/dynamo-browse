package messages

import (
	"context"

	"github.com/lmika/awstools/internal/sqs-browse/models"
)

type MessageSender interface {
	SendMessage(ctx context.Context, msg models.Message, queue string) (string, error)
}
