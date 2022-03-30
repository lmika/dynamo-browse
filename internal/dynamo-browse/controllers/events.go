package controllers

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/awstools/internal/dynamo-browse/models"
)

type NewResultSet struct {
	ResultSet *models.ResultSet
}

func (rs NewResultSet) StatusMessage() string {
	return fmt.Sprintf("%d items returned", len(rs.ResultSet.Items))
}

type SetReadWrite struct {
	NewValue bool
}

type PromptForTableMsg struct {
	Tables     []string
	OnSelected func(tableName string) tea.Cmd
}

type ResultSetUpdated struct{}
