package pasteboardprovider

type NilProvider struct{}

func (n NilProvider) WriteText(bts []byte) error {
	return nil
}
