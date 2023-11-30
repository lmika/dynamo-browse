package controllers

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models"
	"github.com/lmika/dynamo-browse/internal/dynamo-browse/models/relitems"
)

type SetTableItemView struct {
	ViewIndex int
}

type SettingsUpdated struct {
}

type ColumnsUpdated struct {
}

type SetSelectedColumnInColSelector int

type MoveLeftmostDisplayedColumnInTableViewBy int

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

func (rs NewResultSet) RightModeMessage() string {
	var sb strings.Builder

	itemCountStr := applyToN("", len(rs.ResultSet.Items()), "item", "items", "")
	if rs.currentFilter != "" {
		sb.WriteString(fmt.Sprintf("%d of %v", rs.filteredCount, itemCountStr))
	} else {
		sb.WriteString(itemCountStr)
	}

	if !rs.ResultSet.Created.IsZero() {
		sb.WriteString(" • ")
		sb.WriteString(rs.ResultSet.Created.Format(time.Kitchen))
	}

	return sb.String()
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

type ShowColumnOverlay struct{}
type HideColumnOverlay struct{}

type ShowRelatedItemsOverlay struct {
	Items []relitems.RelatedItem
}
type HideRelatedItemsOverlay struct{}
