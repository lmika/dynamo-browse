package inputhistory

type Service struct {
	store HistoryItemStore
}

func New(store HistoryItemStore) *Service {
	return &Service{store: store}
}
