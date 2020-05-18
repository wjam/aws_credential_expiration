package main

import (
	"fmt"
	"os"
	"time"

	"github.com/getlantern/systray"
	"gopkg.in/ini.v1"
)

//go:generate go run generate.go

func main() {
	file, err := credentialsFile()
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(1 * time.Minute)
	systray.Run(ready(file, t), exit(t))
}

func ready(file string, t *time.Ticker) func() {
	return func() {
		updateIconWithExpiration(file)

		select {
		case <-t.C:
			updateIconWithExpiration(file)
		}
	}
}

func updateIconWithExpiration(file string) {
	credentials, err := ini.Load(file)
	if err != nil {
		panic(err)
	}

	status, err := expiringProfiles(credentials)
	if err != nil {
		panic(err)
	}

	if status.HasExpired() {
		systray.SetIcon(redIcon)
	} else if status.HasExpiring() {
		systray.SetIcon(amberIcon)
	} else {
		systray.SetIcon(greenIcon)
	}

	systray.SetTooltip(status.String())
}

func exit(t *time.Ticker) func() {
	return func() {
		t.Stop()
	}
}

func expiringProfiles(credentials *ini.File) (*credentialStatus, error) {
	expired := map[string]time.Time{}
	expiring := map[string]time.Time{}
	for _, section := range credentials.Sections() {
		if section.HasKey("aws_expiration") {
			key, err := section.GetKey("aws_expiration")
			if err != nil {
				return nil, err
			}
			expiration, err := key.TimeFormat(time.RFC3339)
			if err != nil {
				return nil, err
			}

			if expiration.Before(time.Now().Add(-1 * 24 * time.Hour)) {
				continue
			} else if expiration.After(time.Now()) {
				expired[section.Name()] = expiration
			} else if expiration.After(time.Now().Add(10 * time.Minute)) {
				expiring[section.Name()] = expiration
			}
		}
	}

	return &credentialStatus{
		expired:  expired,
		expiring: expiring,
	}, nil
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

type credentialStatus struct {
	expired  map[string]time.Time
	expiring map[string]time.Time
}

func (c *credentialStatus) HasExpiring() bool {
	return len(c.expiring) != 0
}

func (c *credentialStatus) HasExpired() bool {
	return len(c.expired) != 0
}

func (c *credentialStatus) String() string {
	str := ""
	if c.HasExpired() {
		str += "Expired\n"
		for k := range c.expired {
			str += fmt.Sprintf("%s\n", k)
		}
		if c.HasExpiring() {
			str += "\n"
		}
	}

	if c.HasExpiring() {
		str += "Expiring\n"
		for k, v := range c.expiring {
			till := v.Sub(time.Now())
			str += fmt.Sprintf("%s -> %s\n", k, till)
		}
	}

	return str
}
