package jobs

import (
	"testing"
	"time"
)

func TestPolicyDelayGrowsAndIsCapped(t *testing.T) {
	policy := Policy{MaxAttempts: 10, BaseDelay: time.Second, MaxDelay: 8 * time.Second}

	cases := map[int]time.Duration{
		1: time.Second,
		2: 2 * time.Second,
		3: 4 * time.Second,
		4: 8 * time.Second,
		5: 8 * time.Second,
		9: 8 * time.Second,
	}

	for attempt, want := range cases {
		if got := policy.Delay(attempt); got != want {
			t.Errorf("Delay(%d) = %v, want %v", attempt, got, want)
		}
	}
}

func TestDefaultPolicyIsPatientButGivesUp(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MaxAttempts < 3 {
		t.Errorf("MaxAttempts = %d, want enough to ride out a brief outage", policy.MaxAttempts)
	}
	if policy.Delay(1) < time.Second {
		t.Errorf("the first retry is after %v, want long enough not to hammer a remote", policy.Delay(1))
	}
	if policy.Delay(policy.MaxAttempts) > policy.MaxDelay {
		t.Errorf("the last delay %v is above the cap %v", policy.Delay(policy.MaxAttempts), policy.MaxDelay)
	}
}
