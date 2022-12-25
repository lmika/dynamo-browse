package scriptmanager

import (
	"context"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

//go:generate mockery --with-expecter --name UIService
//go:generate mockery --with-expecter --name SessionService

type Ifaces struct {
	UI      UIService
	Session SessionService
}

type UIService interface {
	PrintMessage(ctx context.Context, msg string)

	// Prompt should return a channel which will provide the input from the user.  If the user
	// provides no input, prompt should close the channel without providing anything.
	Prompt(ctx context.Context, msg string) chan string
}

type SessionService interface {
	Query(ctx context.Context, expr string) (*models.ResultSet, error)

	ResultSet() *models.ResultSet
}
