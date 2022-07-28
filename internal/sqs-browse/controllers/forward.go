package controllers

import (
	"context"

	"github.com/lmika/audax/internal/common/ui/uimodels"
	"github.com/lmika/audax/internal/sqs-browse/models"
	"github.com/lmika/audax/internal/sqs-browse/services/messages"
	"github.com/pkg/errors"
)

type MessageSendingController struct {
	messageService *messages.Service
	targetQueue    string
}

func NewMessageSendingController(messageService *messages.Service, targetQueue string) *MessageSendingController {
	return &MessageSendingController{
		messageService: messageService,
		targetQueue:    targetQueue,
	}
}

func (msh *MessageSendingController) ForwardMessage(message models.Message) uimodels.Operation {
	return uimodels.OperationFn(func(ctx context.Context) error {
		uiCtx := uimodels.Ctx(ctx)

		if msh.targetQueue == "" {
			return errors.New("target queue not set")
		}

		messageId, err := msh.messageService.SendTo(ctx, message, msh.targetQueue)
		if err != nil {
			return errors.Wrapf(err, "cannot send message to %v", msh.targetQueue)
		}

		uiCtx.Message("Message sent to " + msh.targetQueue + ", id = " + messageId)
		return nil
	})
}
