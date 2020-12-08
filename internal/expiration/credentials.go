package expiration

import (
	"time"
)

type credentials map[string]profile

func (c credentials) nextExpiration(now time.Time) []time.Duration {
	var nextExpired *time.Time = nil
	var nextExpiring *time.Time = nil
	for _, expiration := range c {
		expired := expiration.expiresAt()
		if !expired.Before(now) && (nextExpired == nil || expired.Before(*nextExpired)) {
			nextExpired = &expired
		}
		expiring := expiration.expiringAt()
		if !expiring.Before(now) && (nextExpiring == nil || expiring.Before(*nextExpiring)) {
			nextExpiring = &expiring
		}
	}

	var times []time.Duration
	if nextExpired != nil {
		times = append(times, nextExpired.Sub(now))
	}
	if nextExpiring != nil {
		times = append(times, nextExpiring.Sub(now))
	}
	return times
}

func (c credentials) groupProfilesByExpiration(now time.Time) (map[string]time.Time, map[string]time.Time, map[string]time.Time) {
	expiredProfiles := map[string]time.Time{}
	expiringProfiles := map[string]time.Time{}
	currentProfiles := map[string]time.Time{}

	for name, expire := range c {
		if expire.expiresAt().Before(now) || expire.expiresAt().Equal(now) {
			expiredProfiles[name] = expire.expiresAt()
		} else if expire.expiringAt().Before(now) {
			expiringProfiles[name] = expire.expiresAt()
		} else {
			currentProfiles[name] = expire.expiresAt()
		}
	}

	return expiredProfiles, expiringProfiles, currentProfiles
}

type profile struct {
	expiration time.Time
}

func (p profile) String() string {
	return p.expiration.String()
}

func (p profile) expiresAt() time.Time {
	return p.expiration
}

func (p profile) expiringAt() time.Time {
	return p.expiration.Add(-1 * 10 * time.Minute)
}
