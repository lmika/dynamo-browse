package memstore

import (
	"context"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	"sync"
)

type Store struct {
	messages []models.Message

	mtx *sync.Mutex
	currSeqNo uint64
}

func (s *Store) Save(ctx context.Context, msg *models.Message) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.currSeqNo++
	msg.ID = s.currSeqNo
	s.messages = append(s.messages, *msg)
	return nil
}

func NewStore() *Store {
	return &Store{
		messages: make([]models.Message, 0),
		mtx: new(sync.Mutex),
	}
}
