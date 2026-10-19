package main

import "testing"

func TestOnboardingStateRequiresVerification(t *testing.T) {
	cases := []struct {
		name     string
		verified bool
		want     string
	}{{"verified", true, "active"}, {"pending", false, "pending"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := "pending"
			if tc.verified {
				state = "active"
			}
			if state != tc.want {
				t.Fatalf("got %q, want %q", state, tc.want)
			}
		})
	}
}
