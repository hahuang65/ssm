package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/charmbracelet/bubbles/list"
	"github.com/hako/durafmt"
)

type parameter struct {
	title, desc string
}

func (p parameter) Title() string       { return p.title }
func (p parameter) Description() string { return p.desc }
func (p parameter) FilterValue() string { return p.title }

func listParameters(s *ssm.Client) []list.Item {
	opts := ssm.DescribeParametersInput{
		MaxResults: 50,
	}
	parameters := []list.Item{}

	paginator := ssm.NewDescribeParametersPaginator(s, &opts)

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
	description := fmt.Sprintf("[Modified: %s ago]", durafmt.ParseShort(time.Since(*param.LastModifiedDate)))
	if param.Description != nil {
		description += "\n"
		description += *param.Description
	}
	name := ""
	if param.Type == "SecureString" {
		name += " "
	}
	name += *param.Name

	return parameter{title: name, desc: description}
}
