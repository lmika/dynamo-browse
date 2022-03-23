package messages

import (
	"context"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	"github.com/pkg/errors"
)

type Service struct {
	messageSender MessageSender
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SendTo(ctx context.Context, msg models.Message, destQueue string) error {
	return errors.Wrapf(s.messageSender.SendMessage(ctx, msg, destQueue), "cannot send message to %v", destQueue)
}
