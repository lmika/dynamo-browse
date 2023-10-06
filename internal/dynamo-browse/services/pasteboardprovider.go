package services

type PasteboardProvider interface {
	ReadText() (string, bool)
	WriteText(bts []byte) error
}
