package settingstore

import (
	"github.com/asdine/storm"
	"github.com/lmika/audax/internal/common/workspaces"
	"github.com/pkg/errors"
	"log"
)

const settingBucket = "Settings"

const (
	keyTableReadOnly     = "ro"
	keyTableDefaultLimit = "default_limit"

	defaultsDefaultLimit = 1000
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
	if err := c.ws.Get(settingBucket, keyTableReadOnly, &b); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return b, err
}

func (c *SettingStore) SetReadOnly(ro bool) error {
	return c.ws.Set(settingBucket, keyTableReadOnly, ro)
}

func (c *SettingStore) DefaultLimit() (limit int) {
	err := c.ws.Get(settingBucket, keyTableDefaultLimit, &limit)
	if err != nil {
		if !errors.Is(err, storm.ErrNotFound) {
			log.Printf("warn: cannot get default limit from workspace, using default value: %v", err)
		}
		return defaultsDefaultLimit
	}
	return limit
}

func (c *SettingStore) SetDefaultLimit(limit int) error {
	return errors.Wrapf(c.ws.Set(settingBucket, keyTableDefaultLimit, &limit), "cannot set default limit to %v", limit)
}
