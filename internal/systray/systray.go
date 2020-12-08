package systray

import "github.com/getlantern/systray"

type Tray struct{}

func NewTray() *Tray {
	return &Tray{}
}

func (t *Tray) SetIcon(b []byte) {
	systray.SetIcon(b)
}

func (t *Tray) SetTooltip(s string) {
	systray.SetTooltip(s)
}
