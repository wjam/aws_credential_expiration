package expiration

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type Systray interface {
	SetIcon([]byte)
	SetTooltip(string)
}

type Expiration struct {
	systray Systray
	now     func() time.Time

	redIcon   []byte
	amberIcon []byte
	greenIcon []byte

	file string
}

func NewExpiration(file string, systray Systray, red []byte, amber []byte, green []byte) Expiration {
	return newExpirationWithTime(file, systray, red, amber, green, time.Now)
}

func newExpirationWithTime(file string, systray Systray, red []byte, amber []byte, green []byte, now func() time.Time) Expiration {
	return Expiration{
		systray:   systray,
		redIcon:   red,
		amberIcon: amber,
		greenIcon: green,
		file:      file,
		now:       now,
	}
}

func (e *Expiration) UpdateIconWithExpiration() error {
	credentials, err := e.loadCredentials()
	if err != nil {
		return err
	}

	status, err := e.expiringProfiles(credentials)
	if err != nil {
		return err
	}

	if status.HasExpired() {
		e.systray.SetIcon(e.redIcon)
	} else if status.HasExpiring() {
		e.systray.SetIcon(e.amberIcon)
	} else {
		e.systray.SetIcon(e.greenIcon)
	}

	e.systray.SetTooltip(status.String())

	return nil
}

func (e *Expiration) loadCredentials() (*ini.File, error) {
	return ini.Load(e.file)
}

func (e *Expiration) expiringProfiles(credentials *ini.File) (*credentialStatus, error) {
	expired := map[string]time.Duration{}
	expiring := map[string]time.Duration{}
	current := map[string]time.Duration{}
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

			now := e.now()
			if expiration.Before(now.Add(-1 * 24 * time.Hour)) {
				// So old to not be of concern
				continue
			} else if expiration.Before(now) {
				expired[section.Name()] = expiration.Sub(now)
			} else if expiration.Before(now.Add(10 * time.Minute)) {
				expiring[section.Name()] = expiration.Sub(now)
			} else {
				current[section.Name()] = expiration.Sub(now)
			}
		}
	}

	return &credentialStatus{
		expired:  expired,
		expiring: expiring,
		current:  current,
	}, nil
}

type credentialStatus struct {
	expired  map[string]time.Duration
	expiring map[string]time.Duration
	current  map[string]time.Duration
}

func (c *credentialStatus) HasCurrent() bool {
	return len(c.current) != 0
}

func (c *credentialStatus) HasExpiring() bool {
	return len(c.expiring) != 0
}

func (c *credentialStatus) HasExpired() bool {
	return len(c.expired) != 0
}

func (c *credentialStatus) String() string {
	var lines []string
	if c.HasExpired() {
		lines = append(lines, "Expired")
		for _, k := range ordered(c.expired) {
			lines = append(lines, k)
		}
		if c.HasExpiring() {
			lines = append(lines, "")
		}
	}

	if c.HasExpiring() {
		lines = append(lines, "Expiring")
		for _, k := range ordered(c.expiring) {
			v := c.expiring[k]
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v))
		}
		if c.HasCurrent() {
			lines = append(lines, "")
		}
	}

	if c.HasCurrent() {
		lines = append(lines, "Current")
		for _, k := range ordered(c.current) {
			v := c.current[k]
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v))
		}
	}

	return strings.Join(lines, "\n")
}

func ordered(m map[string]time.Duration) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
