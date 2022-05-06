package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"git.sr.ht/~hwrd/ssm/internal/parameter"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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
)

type model struct {
	delegateKeys     *delegateKeyMap
	list             list.Model
	loading          bool
	spinner          spinner.Model
	parameterService parameter.Service
}

func newModel(ps parameter.Service) model {
	var (
		delegateKeys = newDelegateKeyMap()
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	delegate := newItemDelegate(delegateKeys, ps)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "AWS SSM"
	l.Styles.Title = titleStyle
	l.StatusMessageLifetime = time.Second * 5

	return model{
		delegateKeys:     delegateKeys,
		list:             l,
		loading:          true,
		spinner:          s,
		parameterService: ps,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.parameterService.List)
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

	case parameter.ListMsg:
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

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}
	ssmClient := ssm.NewFromConfig(cfg)
	p := parameter.Service{SSMClient: ssmClient}

	if len(os.Args[1:]) == 1 {
		// If a single argument is passed in, try to get the value for that key
		fmt.Println(p.Value(os.Args[1]))
	} else {
		if err := tea.NewProgram(newModel(p), tea.WithAltScreen()).Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
