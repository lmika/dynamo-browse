package messages

import (
	"context"

	"github.com/lmika/awstools/internal/sqs-browse/models"
	"github.com/pkg/errors"
)

type Service struct {
	messageSender MessageSender
}

func NewService(messageSender MessageSender) *Service {
	return &Service{
		messageSender: messageSender,
	}
}

func (s *Service) SendTo(ctx context.Context, msg models.Message, destQueue string) (string, error) {
	messageId, err := s.messageSender.SendMessage(ctx, msg, destQueue)
	if err != nil {
		return "", errors.Wrapf(err, "cannot send message to %v", destQueue)
	}
	return messageId, nil
}
