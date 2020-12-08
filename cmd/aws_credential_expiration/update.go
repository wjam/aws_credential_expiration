package main

import (
	"time"

	"github.com/wjam/aws_credential_expiration/internal/systray"

	"github.com/0xAX/notificator"
)

type update struct {
	previousState state
	notify        notify
	tray          tray
}

func newUpdate() *update {
	notify := notificator.New(notificator.Options{
		AppName: "aws_credential_expiration",
	})
	return &update{
		notify: notify,
		tray:   systray.NewTray(),
	}
}

type notify interface {
	Push(title string, text string, iconPath string, urgency string) error
}

type tray interface {
	SetIcon(b []byte)
	SetTooltip(s string)
}

func (u *update) update(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) error {
	toolTop, state := toolTip(expired, expiring, current)
	var icon []byte
	switch state {
	case currentState:
		icon = greenIcon
	case expiringState:
		icon = amberIcon
		if u.previousState != state {
			err := u.notify.Push("Expiration", notifyMessage(expiring, "profile is about to expire", "profiles are about to expire"), "", notificator.UR_NORMAL)
			if err != nil {
				return err
			}
		}
	case expiredState:
		icon = redIcon
		if u.previousState != state {
			err := u.notify.Push("Expiration", notifyMessage(expired, "profile has expired", "profiles have expired"), "", notificator.UR_NORMAL)
			if err != nil {
				return err
			}
		}
	}
	u.tray.SetIcon(icon)
	u.tray.SetTooltip(toolTop)
	u.previousState = state

	return nil
}
