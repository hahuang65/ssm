package tui

import (
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var (
			value string
		)

		if i, ok := m.SelectedItem().(ParameterItem); ok {
			value = i.Value()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.copy):
				clipboard.WriteAll(value)
				return m.NewStatusMessage(statusMessageStyle("Copied ") + valuePreviewStyle(value) + statusMessageStyle(" to clipboard"))

			case key.Matches(msg, keys.preview):
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
