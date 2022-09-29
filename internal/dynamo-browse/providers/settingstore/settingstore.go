package settingstore

import (
	"github.com/asdine/storm"
	"github.com/lmika/audax/internal/common/workspaces"
)

const settingBucket = "Settings"

const (
	keyTableReadOnly = "table_ro"
)

type SettingStore struct {
	ws storm.Node
}

func New(ws *workspaces.Workspace) *SettingStore {
	return &SettingStore{
		ws: ws.DB(),
	}
}

func (c *SettingStore) IsReadOnly() (b bool, err error) {
	err = c.ws.Get(settingBucket, keyTableReadOnly, &b)
	return b, err
}

func (c *SettingStore) SetReadOnly(ro bool) error {
	return c.ws.Set(settingBucket, keyTableReadOnly, ro)
}
