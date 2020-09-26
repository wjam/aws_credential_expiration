package main

import "github.com/getlantern/systray"

type tray struct{}

func (t *tray) SetIcon(b []byte) {
	systray.SetIcon(b)
}

func (t *tray) SetTooltip(s string) {
	systray.SetTooltip(s)
}
