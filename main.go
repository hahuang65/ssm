package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/smithy-go"

	"github.com/hahuang65/ssm/errs"
	"github.com/hahuang65/ssm/parameter"
	"github.com/hahuang65/ssm/tui"
)

// errCredentialsExpired tags errors from SSM that are caused by missing,
// expired, or otherwise unusable AWS credentials.
var errCredentialsExpired = errors.New("AWS credentials are expired or invalid")

// errorMapper holds every translation from a low-level error to a
// user-facing hint. Add a new mapping rather than threading bespoke
// detection through call sites.
var errorMapper = errs.New(
	errs.Mapping{
		Sentinel: errCredentialsExpired,
		Match: smithyAPICode(
			"ExpiredToken",
			"ExpiredTokenException",
			"InvalidClientTokenId",
			"UnrecognizedClientException",
			"MissingAuthenticationTokenException",
			"AuthFailure",
		),
		Message: "Your AWS credentials are expired or invalid. Refresh them and try again.",
	},
)

// smithyAPICode returns a matcher that fires when an error's chain contains
// a smithy.APIError whose code is in codes.
func smithyAPICode(codes ...string) func(error) bool {
	set := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		set[c] = struct{}{}
	}
	return func(err error) bool {
		var ae smithy.APIError
		if !errors.As(err, &ae) {
			return false
		}
		_, ok := set[ae.ErrorCode()]
		return ok
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, errorMapper.Message(err))
	os.Exit(1)
}

func main() {
	c, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		fail(fmt.Errorf("unable to load AWS config: %w", err))
	}

	s := ssm.NewFromConfig(c)
	p := parameter.NewService(s)

	if len(os.Args[1:]) >= 1 {
		// If a single argument is passed in, try to get the value for that key
		key := os.Args[1]
		val, err := p.Get(key)
		if err != nil {
			fail(fmt.Errorf("could not get %q: %w", key, err))
		}

		fmt.Println(val)
	} else {
		err := tui.Start(p)
		if err != nil {
			fail(err)
		}
	}
}
