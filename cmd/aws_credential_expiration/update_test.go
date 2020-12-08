package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate_Update_triggersNotify(t *testing.T) {

	notify := &mockNotify{}
	tray := &mockTray{}

	subject := &update{
		notify:        notify,
		tray:          tray,
		previousState: currentState,
	}

	notify.On("Push", "Expiration", "expired profile has expired", mock.Anything, mock.Anything).Return(nil)
	tray.On("SetIcon", redIcon).Return()
	tray.On("SetTooltip", `Expired
expired

Expiring
expiring -> 2020-02-01 01:01:00 +0000 UTC

Current
current -> 2020-03-01 01:01:00 +0000 UTC`).Return()

	err := subject.update(map[string]time.Time{
		"expired": time.Date(2020, time.January, 1, 1, 1, 0, 0, time.UTC),
	}, map[string]time.Time{
		"expiring": time.Date(2020, time.February, 1, 1, 1, 0, 0, time.UTC),
	}, map[string]time.Time{
		"current": time.Date(2020, time.March, 1, 1, 1, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, expiredState, subject.previousState)

	tray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

func TestUpdate_Update_skipsNotifyIfSameState(t *testing.T) {

	notify := &mockNotify{}
	tray := &mockTray{}

	subject := &update{
		notify:        notify,
		tray:          tray,
		previousState: expiringState,
	}

	tray.On("SetIcon", amberIcon).Return()
	tray.On("SetTooltip", `Expiring
expiring -> 2020-02-01 01:01:00 +0000 UTC

Current
current -> 2020-03-01 01:01:00 +0000 UTC`).Return()

	err := subject.update(map[string]time.Time{}, map[string]time.Time{
		"expiring": time.Date(2020, time.February, 1, 1, 1, 0, 0, time.UTC),
	}, map[string]time.Time{
		"current": time.Date(2020, time.March, 1, 1, 1, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, expiringState, subject.previousState)

	tray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

var _ notify = &mockNotify{}

type mockNotify struct {
	mock.Mock
}

func (m *mockNotify) Push(title string, text string, iconPath string, urgency string) error {
	args := m.Called(title, text, iconPath, urgency)
	return args.Error(0)
}

var _ tray = &mockTray{}

type mockTray struct {
	mock.Mock
}

func (m *mockTray) SetIcon(b []byte) {
	m.Called(b)
}

func (m *mockTray) SetTooltip(s string) {
	m.Called(s)
}
