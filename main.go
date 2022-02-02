package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/charmbracelet/bubbles/key"
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

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type listKeyMap struct {
	// toggleSpinner    key.Binding
	// toggleTitleBar   key.Binding
	// toggleStatusBar  key.Binding
	// togglePagination key.Binding
	// toggleHelpMenu   key.Binding
	// insertItem       key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		// insertItem: key.NewBinding(
		// 	key.WithKeys("a"),
		// 	key.WithHelp("a", "add item"),
		// ),
		// toggleSpinner: key.NewBinding(
		// 	key.WithKeys("s"),
		// 	key.WithHelp("s", "toggle spinner"),
		// ),
		// toggleTitleBar: key.NewBinding(
		// 	key.WithKeys("T"),
		// 	key.WithHelp("T", "toggle title"),
		// ),
		// toggleStatusBar: key.NewBinding(
		// 	key.WithKeys("S"),
		// 	key.WithHelp("S", "toggle status"),
		// ),
		// togglePagination: key.NewBinding(
		// 	key.WithKeys("P"),
		// 	key.WithHelp("P", "toggle pagination"),
		// ),
		// toggleHelpMenu: key.NewBinding(
		// 	key.WithKeys("H"),
		// 	key.WithHelp("H", "toggle help"),
		// ),
	}
}

type model struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
}

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB client
	svc := ssm.NewFromConfig(cfg)
	items := listParameters(svc)

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	parameterList := list.New(items, delegate, 0, 0)
	parameterList.Title = "SSM"
	parameterList.Styles.Title = titleStyle
	parameterList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			// listKeys.toggleSpinner,
			// listKeys.insertItem,
			// listKeys.toggleTitleBar,
			// listKeys.toggleStatusBar,
			// listKeys.togglePagination,
			// listKeys.toggleHelpMenu,
		}
	}

	return model{
		list:         parameterList,
		keys:         listKeys,
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

		switch {
		// case key.Matches(msg, m.keys.toggleSpinner):
		// 	cmd := m.list.ToggleSpinner()
		// 	return m, cmd
		//
		// case key.Matches(msg, m.keys.toggleTitleBar):
		// 	v := !m.list.ShowTitle()
		// 	m.list.SetShowTitle(v)
		// 	m.list.SetShowFilter(v)
		// 	m.list.SetFilteringEnabled(v)
		// 	return m, nil
		//
		// case key.Matches(msg, m.keys.toggleStatusBar):
		// 	m.list.SetShowStatusBar(!m.list.ShowStatusBar())
		// 	return m, nil
		//
		// case key.Matches(msg, m.keys.togglePagination):
		// 	m.list.SetShowPagination(!m.list.ShowPagination())
		// 	return m, nil
		//
		// case key.Matches(msg, m.keys.toggleHelpMenu):
		// 	m.list.SetShowHelp(!m.list.ShowHelp())
		// 	return m, nil
		//
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
