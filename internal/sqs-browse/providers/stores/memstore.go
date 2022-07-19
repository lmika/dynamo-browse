package stores

import (
	"context"
	"github.com/lmika/awstools/internal/common/workspaces"

	"github.com/lmika/awstools/internal/sqs-browse/models"
)

type MessageStore struct {
	ws *workspaces.Workspace
}

func NewMessageStore(ws *workspaces.Workspace) *MessageStore {
	return &MessageStore{ws: ws}
}

func (s *MessageStore) Save(ctx context.Context, msg *models.Message) error {
	return s.ws.DB().Save(msg)
}

func (s *MessageStore) List() (ms []models.Message, err error) {
	//TODO make me better
	err = s.ws.DB().All(&ms)
	return ms, err
}
