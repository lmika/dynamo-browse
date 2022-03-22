package pollmessage

import (
	"context"
	"github.com/lmika/events"
	"github.com/pkg/errors"
	"log"
)

type Service struct {
	store  MessageStore
	poller MessagePoller
	queue  string
	bus    *events.Bus
}

func NewService(store MessageStore, poller MessagePoller, queue string, bus *events.Bus) *Service {
	return &Service{
		store:  store,
		poller: poller,
		queue:  queue,
		bus:    bus,
	}
}

// Poll starts polling for new messages and adding them to the message store
func (s *Service) Poll(ctx context.Context) error {
	for ctx.Err() == nil {
		log.Printf("polling for new messages: %v", s.queue)
		newMsgs, err := s.poller.PollForNewMessages(ctx, s.queue)
		if err != nil {
			return errors.Wrap(err, "unable to poll for messages")
		}

		for _, msg := range newMsgs {
			if err := s.store.Save(ctx, msg); err != nil {
				log.Println("warn: unable to save new message %v", err)
				continue
			}
		}

		s.bus.Fire("new-messages", newMsgs)
	}
	return nil
}
