package expiration

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiration_WatchCredentialsFile_closesOnErrorInUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "temp.ini")
	err := ioutil.WriteFile(path, []byte(`
[initial_expired]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-01T12:31:00.000Z
`), 0600)
	require.NoError(t, err)

	subject := NewExpiration(path, func(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) error {
		return nil
	})

	go func() {
		defer func() {
			err := subject.Close()
			require.NoError(t, err)
		}()
		time.Sleep(1 * time.Second)
		err = ioutil.WriteFile(path, []byte(`
[invalid_expiration_date]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-1T12:50:02.000Z
`), 0600)
		require.NoError(t, err)
		// Wait some time for the parsing to happen and fail rather than trigger the defer block immediately
		time.Sleep(1 * time.Second)
	}()

	err = subject.WatchCredentialsFile()
	assert.Error(t, err)
}

func TestExpiration_WatchCredentialsFile_closesOnErrorInInit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "temp.ini")
	err := ioutil.WriteFile(path, []byte(`
[initial_expired]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-1T12:31:00.000Z
`), 0600)
	require.NoError(t, err)

	subject := NewExpiration(path, func(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) error {
		return nil
	})

	go func() {
		// Wait some time for the parsing to happen and fail rather than trigger the defer block immediately
		time.Sleep(1 * time.Second)
		defer func() {
			err := subject.Close()
			require.NoError(t, err)
		}()
	}()

	err = subject.WatchCredentialsFile()
	assert.Error(t, err)
}

func TestExpiration_WatchCredentialsFile(t *testing.T) {
	now = func() time.Time {
		return time.Date(2020, time.December, 1, 12, 50, 0, 0, time.UTC)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "temp.ini")
	err := ioutil.WriteFile(path, []byte(`
[initial_expired]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-01T12:31:00.000Z
`), 0600)
	require.NoError(t, err)

	var actual []event
	subject := NewExpiration(path, func(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) error {
		actual = append(actual, event{
			expired:  expired,
			expiring: expiring,
			current:  current,
		})
		return nil
	})
	go func() {
		defer func() {
			err := subject.Close()
			require.NoError(t, err)
		}()
		// Update triggered from initial read of credentials file
		assert.Eventually(t, func() bool {
			return containsEventWithExpired(actual, 0, "initial_expired")
		}, 1*time.Second, 10*time.Millisecond)
		err = ioutil.WriteFile(path, []byte(`
[updated_expiring]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-01T12:50:02.000Z

[updated_current]
aws_access_key_id=987654
aws_secret_access_key=2345678
aws_expiration=2020-12-01T13:50:10.000Z
`), 0600)
		require.NoError(t, err)

		// Update triggered from credentials file being updated
		assert.Eventually(t, func() bool {
			return containsEventWithCurrent(actual, 1, "updated_current")
		}, 1*time.Second, 10*time.Millisecond)
		assert.Eventually(t, func() bool {
			return containsEventWithExpiring(actual, 1, "updated_expiring")
		}, 1*time.Second, 10*time.Millisecond)

		// Update triggered from expiring profile becoming expired
		now = func() time.Time {
			return time.Date(2020, time.December, 1, 12, 50, 2, 0, time.UTC)
		}
		assert.Eventually(t, func() bool {
			return containsEventWithExpired(actual, 2, "updated_expiring")
		}, 10*time.Second, 10*time.Millisecond)
	}()

	err = subject.WatchCredentialsFile()
	require.NoError(t, err)

	assert.Len(t, actual, 3)
}

func containsEventWithCurrent(events []event, expectedIndex int, expectedName string) bool {
	if len(events) != expectedIndex+1 {
		return false
	}
	e := events[expectedIndex]
	for name := range e.current {
		if name == expectedName {
			return true
		}
	}
	log.Printf("event current is %s", e)
	return false
}

func containsEventWithExpiring(events []event, expectedIndex int, expectedName string) bool {
	if len(events) != expectedIndex+1 {
		return false
	}
	e := events[expectedIndex]
	for name := range e.expiring {
		if name == expectedName {
			return true
		}
	}
	log.Printf("event expiring is %s", e)
	return false
}

func containsEventWithExpired(events []event, expectedIndex int, expectedName string) bool {
	if len(events) != expectedIndex+1 {
		return false
	}
	e := events[expectedIndex]
	for name := range events[expectedIndex].expired {
		if name == expectedName {
			return true
		}
	}
	log.Printf("event expired is %s", e)
	return false
}

type event struct {
	expired  map[string]time.Time
	expiring map[string]time.Time
	current  map[string]time.Time
}

func (e event) String() string {
	return fmt.Sprintf("current: %s, expiring: %s, expired: %s", e.current, e.expiring, e.expired)
}
