package pasteboardprovider

import (
	"github.com/pkg/errors"
	"golang.design/x/clipboard"
	"sync"
)

type Provider struct {
	mutex         *sync.Mutex
	clipboardInit bool
}

func New() *Provider {
	return &Provider{
		mutex: new(sync.Mutex),
	}
}

func (c *Provider) initClipboard() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.clipboardInit {
		return nil
	}

	if err := clipboard.Init(); err != nil {
		return errors.Wrap(err, "unable to enable clipboard")
	}
	c.clipboardInit = true
	return nil
}

func (c *Provider) WriteText(bts []byte) error {
	if err := c.initClipboard(); err != nil {
		return err
	}

	clipboard.Write(clipboard.FmtText, bts)
	return nil
}
