package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	SSMClient *ssm.Client
)

type model struct {
	list         list.Model
	delegateKeys *delegateKeyMap
}

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
	)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}
	SSMClient = ssm.NewFromConfig(cfg)

	items := listParameters()

	delegate := newItemDelegate(delegateKeys)
	parameterList := list.New(items, delegate, 0, 0)
	parameterList.Title = "SSM"
	parameterList.Styles.Title = titleStyle
	parameterList.StatusMessageLifetime = time.Second * 5

	return model{
		list:         parameterList,
		delegateKeys: delegateKeys,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
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
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func main() {

	if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
