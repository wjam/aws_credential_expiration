package main

import (
	"fmt"
	"os"
	"time"

	"github.com/wjam/aws_credential_expiration/internal/expiration"

	"github.com/getlantern/systray"
)

//go:generate go run generate.go

func main() {
	file, err := credentialsFile()
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(10 * time.Second)
	systray.Run(ready(file, t), exit(t))
}

func ready(file string, t *time.Ticker) func() {
	e := expiration.NewExpiration(file, &tray{}, redIcon, amberIcon, greenIcon)

	return func() {
		err := e.UpdateIconWithExpiration()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case <-t.C:
				err := e.UpdateIconWithExpiration()
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func credentialsFile() (string, error) {
	if file, ok := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE"); ok {
		return file, nil
	}

	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.aws/credentials", home), nil
}

var osUserHomeDir = os.UserHomeDir

func exit(t *time.Ticker) func() {
	return func() {
		t.Stop()
	}
}
