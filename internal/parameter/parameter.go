package parameter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hako/durafmt"
)

type Service struct {
	SSMClient *ssm.Client
}

type Parameter struct {
	description string
	name        string
	title       string
}

type PeekMsg string
type CopyMsg string
type ListMsg []list.Item

func (p Parameter) Title() string       { return p.title }
func (p Parameter) Description() string { return p.description }
func (p Parameter) FilterValue() string { return p.name }

var (
	lastEditedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#9BA92F", Dark: "#9BA92F"}).
		Render
)

func (s Service) List() tea.Msg {
	opts := ssm.DescribeParametersInput{
		MaxResults: 50,
	}
	parameters := []list.Item{}

	paginator := ssm.NewDescribeParametersPaginator(s.SSMClient, &opts)

	for paginator.HasMorePages() {
		res, err := paginator.NextPage(context.TODO())

		if err != nil {
			log.Fatal(err)
		}

		for _, param := range res.Parameters {
			parameters = append(parameters, new(param))
		}
	}

	return ListMsg(parameters)
}

func new(param types.ParameterMetadata) Parameter {
	var (
		name        = *param.Name
		title       string
		description = ""
		lastEdited  = fmt.Sprintf("(Modified: %s ago)", durafmt.ParseShort(time.Since(*param.LastModifiedDate)))
	)

	if param.Description != nil {
		description = *param.Description + " "
	}
	description += lastEditedStyle(lastEdited)

	if param.Type == "SecureString" {
		title = fmt.Sprintf(" %s", name)
	} else {
		title = name
	}

	return Parameter{name: name, title: title, description: description}
}

func (s Service) Peek(name string) tea.Cmd {
	return func() tea.Msg {
		return PeekMsg(s.Value(name))
	}
}

func (s Service) Copy(name string) tea.Cmd {
	return func() tea.Msg {
		return CopyMsg(s.Value(name))
	}
}

func (s Service) Value(name string) string {
	opts := ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: true,
	}

	res, err := s.SSMClient.GetParameter(context.TODO(), &opts)

	if err != nil {
		log.Fatal(err)
	}

	return *res.Parameter.Value
}
