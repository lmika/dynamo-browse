package workspaces

import (
	"github.com/asdine/storm"
	"log"
)

type Workspace struct {
	db *storm.DB
}

func (ws *Workspace) DB() *storm.DB {
	return ws.db
}

func (ws *Workspace) Close() {
	log.Printf("close workspace")
	ws.db.Close()
}
