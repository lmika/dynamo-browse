package statusandprompt

import "github.com/lmika/dynamo-browse/internal/common/ui/events"

type pendingInputState struct {
	originalMsg events.PromptForInputMsg
	historyIdx  int
}

func newPendingInputState(msg events.PromptForInputMsg) *pendingInputState {
	return &pendingInputState{originalMsg: msg, historyIdx: -1}
}
