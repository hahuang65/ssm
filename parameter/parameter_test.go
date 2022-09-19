package parameter_test

import (
	"context"
	"fmt"
	"testing"

	"git.sr.ht/~hwrd/ssm/parameter"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"

	tu "git.sr.ht/~hwrd/ssm/internal/testutils"
)

func parameterService(t *testing.T) parameter.Service {
	t.Helper()
	return parameter.NewService(tu.SSMClient(t))
}

func deleteParameter(t *testing.T, key string) {
	t.Helper()

	tu.SSMClient(t).DeleteParameter(context.TODO(), &ssm.DeleteParameterInput{
		Name: &key,
	})
}

func putParameter(t *testing.T, key string, value string, encrypted bool) {
	t.Helper()

	pt := types.ParameterTypeString
	if encrypted {
		pt = types.ParameterTypeSecureString
	}

	_, err := tu.SSMClient(t).PutParameter(context.TODO(), &ssm.PutParameterInput{
		Name:      &key,
		Value:     &value,
		Type:      pt,
		Overwrite: true, // In tests, we shouldn't error out if a key already exists.
	})
	if err != nil {
		t.Fatalf("Cannot put parameter %q => %q: %v", key, value, err)
	}
}

func TestService(t *testing.T) {
	tu.DockerComposeUp(t)
	ps := parameterService(t)

	t.Run("Get", func(t *testing.T) {
		t.Run("NonExistingKey", func(t *testing.T) {
			key := "foo"
			deleteParameter(t, key)
			got, err := ps.Get(key)

			assert.Equal(t, "", got)
			assert.ErrorContains(t, err, fmt.Sprintf("Parameter %q not found", key))
		})

		t.Run("Key", func(t *testing.T) {
			testCases := []struct {
				name       string
				encryption bool
			}{
				{"Encrypted", true},
				{"Unencrypted", false},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					testCases := []struct {
						key   string
						value string
					}{
						{"foo", "bar"},
						{"/foo/bar/baz", "qux"}, // Leading `/` is required by SSM.
						{"foo.bar.baz", "qux"},
					}

					for _, tc := range testCases {
						t.Run(tc.key, func(t *testing.T) {
							putParameter(t, tc.key, tc.value, true)

							got, err := ps.Get(tc.key)
							assert.Equal(t, tc.value, got)
							assert.Nil(t, err)
						})
					}
				})
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		keyPrefix := "foo"
		valPrefix := "bar"
		count := 100
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("%s%d", keyPrefix, i)
			val := fmt.Sprintf("%s%d", valPrefix, i)
			putParameter(t, key, val, true)
		}

		res, err := ps.List()
		if err != nil {
			t.Fatalf("Could not run `ps.List()`: %v", err)
		}

		parameterMap := make(map[string]string)
		for _, p := range res {
			parameterMap[p.Key] = p.Value
		}

		for i := 0; i < count; i++ {
			key := fmt.Sprintf("%s%d", keyPrefix, i)
			val := fmt.Sprintf("%s%d", valPrefix, i)
			assert.Contains(t, parameterMap, fmt.Sprintf("foo%d", i))
			assert.Equal(t, parameterMap[key], val)
		}
	})
}
