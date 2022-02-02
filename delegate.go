package main

import (
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

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var (
			name  string
			value string
		)

		if i, ok := m.SelectedItem().(parameter); ok {
			name = i.Name()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.copy):
				value = GetParameterValue(name)
				clipboard.WriteAll(value)
				return m.NewStatusMessage(statusMessageStyle("Copied ") + valuePreviewStyle(value) + statusMessageStyle(" to clipboard"))

			case key.Matches(msg, keys.preview):
				value = GetParameterValue(name)
				return m.NewStatusMessage(statusMessageStyle("Peeking at ") + valuePreviewStyle(value))
			}
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
