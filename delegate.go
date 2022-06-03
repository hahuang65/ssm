package main

import (
	"git.sr.ht/~hwrd/ssm/internal/parameter"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#036B46", Dark: "#036B46"}).
				Render

	valuePreviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

func newItemDelegate(keys *delegateKeyMap, p parameter.Service) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var (
			name  string
			value string
		)

		if i, ok := m.SelectedItem().(parameter.Parameter); ok {
			name = i.FilterValue()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.copy):
				return p.Copy(name)

			case key.Matches(msg, keys.preview):
				return p.Peek(name)
			}

		case parameter.CopyMsg:
			value = string(msg)
			clipboard.WriteAll(value)
			return m.NewStatusMessage(statusMessageStyle("Copied ") + valuePreviewStyle(value) + statusMessageStyle(" to clipboard"))

		case parameter.PeekMsg:
			value = string(msg)
			return m.NewStatusMessage(statusMessageStyle("Peeking at ") + valuePreviewStyle(value))
		}

		return nil
	}

	help := []key.Binding{keys.copy, keys.preview}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	copy    key.Binding
	preview key.Binding
}

func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.copy,
		d.preview,
	}
}

func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.copy,
			d.preview,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		copy: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "copy value"),
		),
		preview: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "preview"),
		),
	}
}
