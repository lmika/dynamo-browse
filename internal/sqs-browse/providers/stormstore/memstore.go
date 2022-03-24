package stormstore

import (
	"context"
	"github.com/asdine/storm"
	"github.com/lmika/awstools/internal/sqs-browse/models"
	"github.com/pkg/errors"
)

type Store struct {
	db *storm.DB
}

// TODO: should probably be a workspace provider
func NewStore(filename string) (*Store, error) {
	db, err := storm.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open store %v", filename)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Save(ctx context.Context, msg *models.Message) error {
	return s.db.Save(msg)
}
