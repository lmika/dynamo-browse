package services

type HistoryProvider interface {
	// Len returns the number of historical items
	Len() int

	// Item returns the historical item at index 'idx', where items are chronologically ordered such that the
	// item at 0 is the oldest item.
	Item(idx int) string

	// PutItem adds an item to the history
	PutItem(item string)
}
