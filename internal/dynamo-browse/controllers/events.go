package controllers

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/audax/internal/dynamo-browse/models"
)

type SetTableItemView struct {
	ViewIndex int
}

type SettingsUpdated struct {
}

type NewResultSet struct {
	ResultSet     *models.ResultSet
	currentFilter string
	filteredCount int
	statusMessage string
}

func (rs NewResultSet) ModeMessage() string {
	var modeLine string

	if rs.ResultSet.Query != nil {
		modeLine = rs.ResultSet.Query.String()
	} else {
		modeLine = "All results"
	}

	if rs.currentFilter != "" {
		modeLine = fmt.Sprintf("%v - Filter: '%v'", modeLine, rs.currentFilter)
	}
	return modeLine
}

func (rs NewResultSet) StatusMessage() string {
	if rs.statusMessage != "" {
		return rs.statusMessage
	}

	if rs.currentFilter != "" {
		return fmt.Sprintf("%d of %d items returned", rs.filteredCount, len(rs.ResultSet.Items()))
	} else {
		return fmt.Sprintf("%d items returned", len(rs.ResultSet.Items()))
	}
}

type PromptForTableMsg struct {
	Tables     []string
	OnSelected func(tableName string) tea.Msg
}

type ResultSetUpdated struct {
	statusMessage string
}

func (rs ResultSetUpdated) StatusMessage() string {
	return rs.statusMessage
}
