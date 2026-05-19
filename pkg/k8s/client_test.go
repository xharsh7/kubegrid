package k8s

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 3 * time.Hour, "3h"},
		{"days", 50 * time.Hour, "2d"},
		{"edge just under minute", 59 * time.Second, "59s"},
		{"edge exactly minute", 60 * time.Second, "1m"},
		{"edge just under hour", 59 * time.Minute, "59m"},
		{"edge exactly hour", 60 * time.Minute, "1h"},
		{"edge just under day", 23 * time.Hour, "23h"},
		{"edge exactly day", 24 * time.Hour, "1d"},
		{"zero duration", 0, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAge(tt.duration); got != tt.want {
				t.Errorf("FormatAge(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}
