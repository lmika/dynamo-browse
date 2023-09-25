package settingstore

import (
	"github.com/asdine/storm"
	"github.com/lmika/dynamo-browse/internal/common/workspaces"
	"github.com/pkg/errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const settingBucket = "Settings"

const (
	keyTableReadOnly     = "ro"
	keyTableDefaultLimit = "default_limit"
	keyScriptLookupPath  = "script_lookup_path"

	defaultsDefaultLimit     = 1000
	defaultScriptLookupPaths = "${HOME}/.config/audax/dynamo-browse/scripts"
)

type SettingStore struct {
	ws storm.Node
}

func New(ws *workspaces.Workspace) *SettingStore {
	return &SettingStore{
		ws: ws.DB(),
	}
}

func (c *SettingStore) SetScriptLookupPaths(value string) error {
	return c.ws.Set(settingBucket, keyTableReadOnly, value)
}

func (c *SettingStore) ScriptLookupPaths() string {
	paths, err := c.getStringValue(keyScriptLookupPath, defaultScriptLookupPaths)
	if err != nil {
		return ""
	}
	return paths
}

func (c *SettingStore) ScriptLookupFS() ([]fs.FS, error) {
	paths, err := c.getStringValue(keyScriptLookupPath, defaultScriptLookupPaths)
	if err != nil {
		return nil, err
	}

	if paths == "" {
		return nil, nil
	}

	fs := make([]fs.FS, 0, len(paths))
	for _, path := range strings.Split(paths, string(os.PathListSeparator)) {
		expandedPath := os.ExpandEnv(path)

		var absPath string
		if filepath.IsAbs(path) {
			absPath = expandedPath
		} else {
			absPath, err = filepath.Abs(expandedPath)
			if err != nil {
				log.Printf("warn: cannot include script lookup path '%v': %v", expandedPath, err)
				continue
			}
		}

		if stat, err := os.Stat(absPath); err != nil {
			log.Printf("warn: cannot stat script lookup path '%v': %v", expandedPath, err)
		} else if stat.IsDir() {
			log.Printf("warn: script lookup path '%v' is not a directory", expandedPath)
		}

		log.Printf("adding script lookup path: %v", absPath)
		fs = append(fs, os.DirFS(absPath))
	}

	return fs, nil
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

func (c *SettingStore) getStringValue(key string, def string) (string, error) {
	var val string
	if err := c.ws.Get(settingBucket, key, &val); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return def, nil
		}
		return "", err
	}
	return val, nil
}
