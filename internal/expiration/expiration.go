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

type Notify interface {
	Push(string) error
}

type Expiration struct {
	systray Systray
	notify  Notify
	now     func() time.Time

	redIcon   []byte
	amberIcon []byte
	greenIcon []byte

	obsolete time.Duration
	expiring time.Duration

	previous state

	file string
}

type state int

const (
	currentState state = iota
	expiringState
	expiredState
)

func NewExpiration(file string, systray Systray, notify Notify, red []byte, amber []byte, green []byte) Expiration {
	return newExpirationWithTime(file, systray, notify, red, amber, green, time.Now)
}

func newExpirationWithTime(file string, systray Systray, notify Notify, red []byte, amber []byte, green []byte, now func() time.Time) Expiration {
	return Expiration{
		systray:   systray,
		notify:    notify,
		redIcon:   red,
		amberIcon: amber,
		greenIcon: green,
		file:      file,
		now:       now,
		obsolete:  -1 * 8 * time.Hour,
		expiring:  10 * time.Minute,
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
		if e.previous != expiredState {
			if err := e.notify.Push(status.Expired()); err != nil {
				return err
			}
			e.previous = expiredState
		}
	} else if status.HasExpiring() {
		e.systray.SetIcon(e.amberIcon)
		if e.previous != expiringState {
			if err := e.notify.Push(status.Expiring()); err != nil {
				return err
			}
			e.previous = expiringState
		}
	} else {
		e.systray.SetIcon(e.greenIcon)
		e.previous = currentState
	}

	e.systray.SetTooltip(status.ToolTip())

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
			if expiration.Before(now.Add(e.obsolete)) {
				// So old to not be of concern
				continue
			} else if expiration.Before(now) {
				expired[section.Name()] = expiration.Sub(now)
			} else if expiration.Before(now.Add(e.expiring)) {
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

func (c credentialStatus) Expired() string {
	var profiles []string
	for name := range c.expired {
		profiles = append(profiles, name)
	}

	return notifyMessage(profiles, "profile has expired", "profiles have expired")
}

func (c credentialStatus) Expiring() string {
	var profiles []string
	for name := range c.expiring {
		profiles = append(profiles, name)
	}

	return notifyMessage(profiles, "profile is about to expire", "profiles are about to expire")
}

func (c *credentialStatus) ToolTip() string {
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
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v.Truncate(time.Second)))
		}
		if c.HasCurrent() {
			lines = append(lines, "")
		}
	}

	if c.HasCurrent() {
		lines = append(lines, "Current")
		for _, k := range ordered(c.current) {
			v := c.current[k]
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v.Truncate(time.Second)))
		}
	}

	return strings.Join(lines, "\n")
}

func notifyMessage(profiles []string, singular string, plural string) string {
	if len(profiles) > 1 {
		return fmt.Sprintf("%s %s", concat(profiles), plural)
	}
	return fmt.Sprintf("%s %s", concat(profiles), singular)
}

func concat(parts []string) string {
	sort.Strings(parts)

	s := new(strings.Builder)
	for i, part := range parts {
		if s.Len() != 0 {
			if i == len(parts)-1 {
				s.WriteString(" and ")
			} else {
				s.WriteString(", ")
			}
		}
		s.WriteString(part)
	}
	return s.String()
}

func ordered(m map[string]time.Duration) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
