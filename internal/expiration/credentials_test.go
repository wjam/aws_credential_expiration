package expiration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCredentials_groupProfilesByExpiration(t *testing.T) {
	now := time.Now()

	currentTime := now.Add(11 * time.Minute)
	expiringTime := now.Add(9 * time.Minute)
	expiredTime := now.Add(-1 * time.Minute)
	subject := credentials{
		"current":  profile{currentTime},
		"expiring": profile{expiringTime},
		"expired":  profile{expiredTime},
	}

	expired, expiring, current := subject.groupProfilesByExpiration(now)
	assert.Equal(t, map[string]time.Time{
		"expired": expiredTime,
	}, expired)
	assert.Equal(t, map[string]time.Time{
		"expiring": expiringTime,
	}, expiring)
	assert.Equal(t, map[string]time.Time{
		"current": currentTime,
	}, current)
}
