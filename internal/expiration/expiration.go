package expiration

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/ini.v1"
)

func NewExpiration(path string, update Update) *Expiration {
	return &Expiration{
		path:   path,
		update: update,
		fin:    make(chan error, 1),
	}
}

type Update func(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) error

type Expiration struct {
	path string
	fin  chan error

	creds  credentials
	update Update
}

func (e *Expiration) WatchCredentialsFile() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = w.Add(e.path)
	if err != nil {
		return err
	}

	closeChan := make(chan bool, 1)

	go func() {

		// Trigger once to ensure the tooltip is updated
		timers, err := e.nextEvent(nil)
		if err != nil {
			e.fin <- err
			return
		}

		for {
			select {
			case event := <-w.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					timers, err = e.nextEvent(timers)

					if err != nil {
						e.fin <- err
						return
					}
				}
			case <-closeChan:
				for _, timer := range timers {
					timer.Stop()
				}
				_ = w.Close()
				return
			}
		}
	}()
	select {
	case err := <-e.fin:
		closeChan <- true
		return err
	}
}

func (e Expiration) Close() error {
	e.fin <- nil
	return nil
}

func (e Expiration) nextEvent(previous []Stoppable) ([]Stoppable, error) {
	err := e.updateCredentials()
	if err != nil {
		return nil, err
	}

	durs := e.nextEvents()

	e.triggerUpdate()

	var timers []Stoppable
	for _, timer := range durs {
		log.Printf("Event will trigger after %s", timer)
		timers = append(timers, time.AfterFunc(timer, e.triggerUpdate))
	}

	// Make sure the previous timers can be GCed
	for _, timer := range previous {
		timer.Stop()
	}

	return timers, nil
}

func (e *Expiration) triggerUpdate() {
	expired, expiring, current := e.creds.groupProfilesByExpiration(now())
	err := e.update(expired, expiring, current)
	if err != nil {
		e.fin <- err
	}
}

func (e *Expiration) updateCredentials() error {
	f, err := ini.Load(e.path)
	if err != nil {
		return err
	}
	creds := credentials{}
	for _, section := range f.Sections() {
		if section.HasKey("aws_expiration") {
			key, err := section.GetKey("aws_expiration")
			if err != nil {
				return err
			}
			expiration, err := key.TimeFormat(time.RFC3339)
			if err != nil {
				return err
			}

			creds[section.Name()] = profile{expiration}
		}
	}

	e.creds = creds
	log.Printf("Updated credentials: %v", e.creds)
	return nil
}

func (e Expiration) nextEvents() []time.Duration {
	return e.creds.nextExpiration(now())
}

var now = time.Now

type Stoppable interface {
	Stop() bool
}
