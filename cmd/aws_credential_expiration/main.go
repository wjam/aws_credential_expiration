package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/wjam/aws_credential_expiration/internal/expiration"
)

//go:embed red.png
var redIcon []byte

//go:embed amber.png
var amberIcon []byte

//go:embed green.png
var greenIcon []byte

func main() {
	file, err := credentialsFile()
	if err != nil {
		panic(err)
	}

	u := newUpdate()

	e := expiration.NewExpiration(file, u.update)

	systray.Run(func() {
		if err := e.WatchCredentialsFile(); err != nil {
			panic(err)
		}
	}, func() {
		if err := e.Close(); err != nil {
			panic(err)
		}
	})
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
