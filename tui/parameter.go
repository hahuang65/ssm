package tui

import (
	"fmt"
	"log"
	"time"

	"git.sr.ht/~hwrd/ssm/parameter"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hako/durafmt"
)

type ListMsg []list.Item

type ParameterItem struct {
	param parameter.Parameter
}

func (p ParameterItem) Title() string {
	title := p.param.Key
	if p.param.Type == "SecureString" {
		title = fmt.Sprintf(" %s", p.param.Key)
	}
	return title
}

func (p ParameterItem) LastEdited() string {
	relative := durafmt.ParseShort(time.Since(*p.param.LastModified))
	return fmt.Sprintf("(Modified: %s ago)", relative)
}

func (p ParameterItem) Description() string { return p.param.Description }
func (p ParameterItem) FilterValue() string { return p.param.Key }
func (p ParameterItem) Value() string       { return p.param.Value }

func (m model) listParameters() tea.Msg {
	items := []list.Item{}

	params, err := m.parameterService.List()
	if err != nil {
		log.Fatalf("Could not list parameters from SSM: %v", err)
	}

	for _, p := range params {
		items = append(items, ParameterItem{param: p})
	}

	return ListMsg(items)
}
