package controllers

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/models"
	"io/fs"
)

type TableReadService interface {
	ListTables(background context.Context) ([]string, error)
	Describe(ctx context.Context, table string) (*models.TableInfo, error)
	Scan(ctx context.Context, tableInfo *models.TableInfo) (*models.ResultSet, error)
	Filter(resultSet *models.ResultSet, filter string) *models.ResultSet
	ScanOrQuery(ctx context.Context, tableInfo *models.TableInfo, query models.Queryable, exclusiveStartKey map[string]types.AttributeValue) (*models.ResultSet, error)
	NextPage(ctx context.Context, resultSet *models.ResultSet) (*models.ResultSet, error)
}

type SettingsProvider interface {
	IsReadOnly() (bool, error)
	SetReadOnly(ro bool) error
	DefaultLimit() (limit int)
	SetDefaultLimit(limit int) error
	ScriptLookupFS() ([]fs.FS, error)
	SetScriptLookupPaths(value string) error
	ScriptLookupPaths() string
}

type CustomKeyBindingSource interface {
	LookupBinding(theKey string) string
	CustomKeyCommand(key string) tea.Cmd
	UnbindKey(key string)
	Rebind(bindingName string, newKey string) error
}
