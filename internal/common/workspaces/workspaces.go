package workspaces

import (
	"github.com/asdine/storm"
)

type Workspace struct {
	db *storm.DB
}

func (ws *Workspace) DB() *storm.DB {
	return ws.db
}

func (ws *Workspace) Close() {
	ws.db.Close()
}
