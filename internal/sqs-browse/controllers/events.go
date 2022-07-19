package controllers

import "github.com/lmika/awstools/internal/sqs-browse/models"

type MessageListUpdated struct {
	MessageList models.MessageList
}

type EditMessage struct {
}
