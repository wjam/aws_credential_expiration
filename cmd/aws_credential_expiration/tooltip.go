package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type state int

const (
	currentState state = iota
	expiringState
	expiredState
)

func notifyMessage(profiles map[string]time.Time, singular string, plural string) string {
	if len(profiles) > 1 {
		return fmt.Sprintf("%s %s", concat(profiles), plural)
	}
	return fmt.Sprintf("%s %s", concat(profiles), singular)
}

func concat(profiles map[string]time.Time) string {

	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}

	sort.Strings(names)

	s := new(strings.Builder)
	for i, name := range names {
		if s.Len() != 0 {
			if i == len(names)-1 {
				s.WriteString(" and ")
			} else {
				s.WriteString(", ")
			}
		}
		s.WriteString(name)
	}
	return s.String()
}

func toolTip(expired map[string]time.Time, expiring map[string]time.Time, current map[string]time.Time) (string, state) {
	var lines []string

	s := currentState
	if len(expired) != 0 {
		lines = append(lines, "Expired")
		for _, k := range orderedTime(expired) {
			lines = append(lines, k)
		}
		if len(expiring) != 0 {
			lines = append(lines, "")
		}
		s = expiredState
	}

	if len(expiring) != 0 {
		lines = append(lines, "Expiring")
		for _, k := range orderedTime(expiring) {
			v := expiring[k]
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v.Truncate(time.Second)))
		}
		if len(current) != 0 {
			lines = append(lines, "")
		}
		if s < expiringState {
			s = expiringState
		}
	}

	if len(current) != 0 {
		lines = append(lines, "Current")
		for _, k := range orderedTime(current) {
			v := current[k]
			lines = append(lines, fmt.Sprintf("%s -> %s", k, v.Truncate(time.Second)))
		}
	}

	return strings.Join(lines, "\n"), s
}

func orderedTime(m map[string]time.Time) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
