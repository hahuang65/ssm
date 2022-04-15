package main

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

type parameter struct {
	description string
	name        string
	title       string
}

type PeekParameterMsg string
type CopyParameterMsg string

func (p parameter) Title() string       { return p.title }
func (p parameter) Name() string        { return p.name }
func (p parameter) Description() string { return p.description }
func (p parameter) FilterValue() string { return p.name }

var (
	lastEditedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#9BA92F", Dark: "#9BA92F"}).
		Render
)

func listParameters() []list.Item {
	opts := ssm.DescribeParametersInput{
		MaxResults: 50,
	}
	parameters := []list.Item{}

	paginator := ssm.NewDescribeParametersPaginator(SSMClient, &opts)

	for paginator.HasMorePages() {
		res, err := paginator.NextPage(context.TODO())

		if err != nil {
			log.Fatal(err)
		}

		for _, param := range res.Parameters {
			parameters = append(parameters, NewParameterItem(param))
		}
	}

	return parameters
}

func NewParameterItem(param types.ParameterMetadata) parameter {
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

	return parameter{name: name, title: title, description: description}
}

func PeekParameter(name string) tea.Cmd {
	return func() tea.Msg {
		return PeekParameterMsg(getParameterValue(name))
	}
}

func CopyParameter(name string) tea.Cmd {
	return func() tea.Msg {
		return CopyParameterMsg(getParameterValue(name))
	}
}

func getParameterValue(name string) string {
	opts := ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: true,
	}

	res, err := SSMClient.GetParameter(context.TODO(), &opts)

	if err != nil {
		log.Fatal(err)
	}

	return *res.Parameter.Value
}
