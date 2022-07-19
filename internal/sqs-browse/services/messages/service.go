package messages

import (
	"context"

	"github.com/lmika/awstools/internal/sqs-browse/models"
	"github.com/pkg/errors"
)

type Service struct {
	messageSender MessageSender
	store         MessageStore
}

func NewService(store MessageStore, messageSender MessageSender) *Service {
	return &Service{
		store:         store,
		messageSender: messageSender,
	}
}

func (s *Service) List() ([]models.Message, error) {
	return s.store.List()
}

func (s *Service) SendTo(ctx context.Context, msg models.Message, destQueue string) (string, error) {
	messageId, err := s.messageSender.SendMessage(ctx, msg, destQueue)
	if err != nil {
		return "", errors.Wrapf(err, "cannot send message to %v", destQueue)
	}
	return messageId, nil
}
