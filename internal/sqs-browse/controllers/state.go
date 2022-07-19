package controllers

import "github.com/lmika/awstools/internal/sqs-browse/models"

type State struct {
	messageList models.MessageList
}

func (s *State) MessageList() models.MessageList {
	return s.messageList
}
