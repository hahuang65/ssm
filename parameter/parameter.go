package parameter

import (
	"context"
	"errors"
	"fmt"

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
