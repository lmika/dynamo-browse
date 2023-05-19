package services

type PasteboardProvider interface {
	WriteText(bts []byte) error
}
