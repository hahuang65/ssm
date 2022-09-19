package parameter

import (
	"context"
	"errors"
	"fmt"
	"time"

	ssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type Service struct {
	client *ssm.Client
}

type ParameterNotFound struct {
	name string
	Err  error
}

type Parameter struct {
	Description  string
	Key          string
	Value        string
	Type         types.ParameterType
	LastModified *time.Time
}

func (e ParameterNotFound) Error() string {
	return fmt.Sprintf("Parameter %q not found.", e.name)
}

func NewService(c *ssm.Client) Service {
	return Service{client: c}
}

func (s Service) Get(key string) (string, error) {
	opts := ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: true,
	}

	res, err := s.client.GetParameter(context.TODO(), &opts)
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return "", ParameterNotFound{name: key, Err: err}
		} else {
			return "", err
		}
	}

	return *res.Parameter.Value, nil
}

func (s Service) List() ([]Parameter, error) {
	descOpts := ssm.DescribeParametersInput{
		MaxResults: 10,
	}
	ret := []Parameter{}

	paginator := ssm.NewDescribeParametersPaginator(s.client, &descOpts)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []Parameter{}, err
		}

		params, err := s.parametersFromPage(page)
		ret = append(ret, params...)
	}

	return ret, nil
}

func (s Service) parametersFromPage(page *ssm.DescribeParametersOutput) ([]Parameter, error) {
	ret := []Parameter{}
	names := []string{}
	descriptions := make(map[string]string)

	for _, p := range page.Parameters {
		names = append(names, *p.Name)
		description := ""
		if p.Description != nil {
			description = *p.Description
		}
		descriptions[*p.Name] = description
	}

	getOpts := ssm.GetParametersInput{
		Names:          names,
		WithDecryption: true,
	}

	res, err := s.client.GetParameters(context.TODO(), &getOpts)
	if err != nil {
		return []Parameter{}, err
	}

	for _, p := range res.Parameters {
		ret = append(ret, Parameter{
			Key:          *p.Name,
			Value:        *p.Value,
			Description:  descriptions[*p.Name],
			Type:         p.Type,
			LastModified: p.LastModifiedDate,
		})
	}

	return ret, nil
}
