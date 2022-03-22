package pollmessage

import (
	"context"
	"github.com/lmika/awstools/internal/sqs-browse/models"
)

type MessageStore interface {
	Save(ctx context.Context, msg *models.Message) error
}

type MessagePoller interface {
	PollForNewMessages(ctx context.Context, queue string) ([]*models.Message, error)
}