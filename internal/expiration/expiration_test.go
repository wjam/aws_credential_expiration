package expiration

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConcat_joinsItemsCorrectly(t *testing.T) {
	actual := concat([]string{"one", "two", "four"})
	assert.Equal(t, "four, one and two", actual)
}

func TestExpiration_notificationOnlySentFirstTime(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "credentials")
	err := ioutil.WriteFile(file, []byte(`
[prod]
aws_access_key_id=123456
aws_secret_access_key=8765432
foo=bar
aws_expiration=2020-09-26T16:31:59.000Z
`), 0644)
	require.NoError(t, err)

	systray := new(mockedSystray)
	notify := new(mockNotify)
	subject := newExpirationWithTime(
		file,
		systray,
		notify,
		red,
		amber,
		green,
		constantTime(time.Date(2020, 9, 26, 16, 22, 0, 0, time.UTC)),
	)

	systray.Test(t)
	systray.On("SetIcon", mock.Anything).Return()
	systray.On("SetTooltip", mock.Anything).Return()

	subject.previous = expiringState
	err = subject.UpdateIconWithExpiration()
	require.NoError(t, err)

	systray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

func TestExpiration_UpdateIconWithExpiration_onlyExpiring(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "credentials")
	err := ioutil.WriteFile(file, []byte(`
[prod]
aws_access_key_id=123456
aws_secret_access_key=8765432
foo=bar
aws_expiration=2020-09-26T16:31:59.000Z

[uat]
aws_access_key_id=asdfg
aws_secret_access_key=jhgfd
aws_expiration=2020-09-26T16:22:01.000Z

[dev]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-09-27T16:31:59.000Z
`), 0644)
	require.NoError(t, err)

	systray := new(mockedSystray)
	notify := new(mockNotify)
	subject := newExpirationWithTime(
		file,
		systray,
		notify,
		red,
		amber,
		green,
		constantTime(time.Date(2020, 9, 26, 16, 22, 0, 0, time.UTC)),
	)

	systray.Test(t)
	systray.On("SetIcon", amber).Return()
	systray.On("SetTooltip", `Expiring
prod -> 9m59s
uat -> 1s

Current
dev -> 24h9m59s`).Return()

	notify.Test(t)
	notify.On("Push", "prod and uat profiles are about to expire").Return(nil)

	err = subject.UpdateIconWithExpiration()
	require.NoError(t, err)
	assert.Equal(t, expiringState, subject.previous)

	systray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

func TestExpiration_UpdateIconWithExpiration_expiringAndExpired(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "credentials")
	err := ioutil.WriteFile(file, []byte(`
[prod]
aws_access_key_id=123456
aws_secret_access_key=8765432
foo=bar
aws_expiration=2020-09-26T16:47:59.000Z

[dev]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-09-26T16:31:59.000Z
`), 0644)
	require.NoError(t, err)

	systray := new(mockedSystray)
	notify := new(mockNotify)
	subject := newExpirationWithTime(
		file,
		systray,
		notify,
		red,
		amber,
		green,
		constantTime(time.Date(2020, 9, 26, 16, 45, 0, 0, time.UTC)),
	)

	systray.Test(t)
	systray.On("SetIcon", red).Return()
	systray.On("SetTooltip", `Expired
dev

Expiring
prod -> 2m59s`).Return()

	notify.Test(t)
	notify.On("Push", "dev profile has expired").Return(nil)

	err = subject.UpdateIconWithExpiration()
	require.NoError(t, err)
	assert.Equal(t, expiredState, subject.previous)

	systray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

func TestExpiration_UpdateIconWithExpiration_allCurrent(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "credentials")
	err := ioutil.WriteFile(file, []byte(`
[prod]
aws_access_key_id=123456
aws_secret_access_key=8765432
foo=bar
aws_expiration=2020-09-25T16:44:59.250Z

[uat]
aws_access_key_id=asdfg
aws_secret_access_key=jhgfd
aws_expiration=2020-09-26T16:56:01.100Z

[dev]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-09-27T16:31:59.300Z
`), 0644)
	require.NoError(t, err)

	systray := new(mockedSystray)
	notify := new(mockNotify)
	subject := newExpirationWithTime(
		file,
		systray,
		notify,
		red,
		amber,
		green,
		constantTime(time.Date(2020, 9, 26, 16, 45, 0, 0, time.UTC)),
	)

	systray.Test(t)
	systray.On("SetIcon", green).Return()
	systray.On("SetTooltip", `Current
dev -> 23h46m59s
uat -> 11m1s`).Return()

	err = subject.UpdateIconWithExpiration()
	require.NoError(t, err)
	assert.Equal(t, currentState, subject.previous)

	systray.AssertExpectations(t)
	notify.AssertExpectations(t)
}

var red = []byte{0x01}
var amber = []byte{0x02}
var green = []byte{0x03}

var _ Systray = &mockedSystray{}

type mockedSystray struct {
	mock.Mock
}

func (m *mockedSystray) SetIcon(bytes []byte) {
	m.Called(bytes)
}

func (m *mockedSystray) SetTooltip(s string) {
	m.Called(s)
}

var _ Notify = &mockNotify{}

type mockNotify struct {
	mock.Mock
}

func (m *mockNotify) Push(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func constantTime(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}
