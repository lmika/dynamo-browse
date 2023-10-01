package pasteboardprovider

type NilProvider struct{}

func (NilProvider) ReadText() (string, bool) {
	return "", false
}

func (n NilProvider) WriteText(bts []byte) error {
	return nil
}
