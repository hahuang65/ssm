package main

import (
	"context"
	"fmt"
	"log"
	"os"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"git.sr.ht/~hwrd/ssm/parameter"
	"git.sr.ht/~hwrd/ssm/tui"
)

func main() {
	c, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
	}

	s := ssm.NewFromConfig(c)
	p := parameter.NewService(s)

	if len(os.Args[1:]) >= 1 {
		// If a single argument is passed in, try to get the value for that key
		key := os.Args[1]
		val, err := p.Get(key)
		if err != nil {
			log.Fatalf("Could not get %q: %v", key, err)
		}

		fmt.Println(val)
	} else {
		err := tui.Start(p)
		if err != nil {
			log.Fatalf("Could not start TUI: %v", err)
		}
	}
}
