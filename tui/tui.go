package tui

import (
	"fmt"
	"time"

	"git.sr.ht/~hwrd/ssm/parameter"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#036B46")).
			Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#036B46", Dark: "#036B46"}).
				Render

	valuePreviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type model struct {
	delegateKeys     *delegateKeyMap
	list             list.Model
	loading          bool
	spinner          spinner.Model
	parameterService parameter.Service
}

func newModel(p parameter.Service) model {
	delegateKeys := newDelegateKeyMap()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New([]list.Item{}, newItemDelegate(delegateKeys), 0, 0)
	l.Title = "AWS SSM"
	l.Styles.Title = titleStyle
	l.StatusMessageLifetime = time.Second * 5

	return model{
		delegateKeys:     delegateKeys,
		list:             l,
		loading:          true,
		spinner:          s,
		parameterService: p,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.listParameters)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

	case ListMsg:
		m.list.SetItems(msg)
		m.loading = false

	case spinner.TickMsg:
		if m.loading {
			newSpinner, cmd := m.spinner.Update(msg)
			m.spinner = newSpinner
			cmds = append(cmds, cmd)
		}
	}

	// This will also call our delegate's update function.
	newList, cmd := m.list.Update(msg)
	m.list = newList
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("\n\n   %s Loading SSM parameters\n\n", m.spinner.View())
	} else {
		return appStyle.Render(m.list.View())
	}
}

func Start(p parameter.Service) error {
	return tea.NewProgram(newModel(p), tea.WithAltScreen()).Start()
}
