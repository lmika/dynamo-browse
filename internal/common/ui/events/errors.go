package events

import (
	"github.com/lmika/awstools/internal/common/ui/uimodels"
)

// Error indicates that an error occurred
type Error error

// Message indicates that a message should be shown to the user
type Message string

// PromptForInput indicates that the context is requesting a line of input
type PromptForInput struct {
	Prompt string
	OnDone uimodels.Operation
}