package scriptmanager

import "context"

//go:generate mockery --with-expecter --name UIService

type Ifaces struct {
	UI UIService
}

type UIService interface {
	PrintMessage(ctx context.Context, msg string)

	// Prompt should return a channel which will provide the input from the user.  If the user
	// provides no input, prompt should close the channel without providing anything.
	Prompt(ctx context.Context, msg string) chan string
}
