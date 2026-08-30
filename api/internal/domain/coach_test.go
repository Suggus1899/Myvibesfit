package domain

import (
	"testing"
	"time"
)

func TestNeedsAttention(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) *time.Time { t := now.Add(-d); return &t }

	tests := []struct {
		name          string
		hasAssignment bool
		lastSession   *time.Time
		want          bool
	}{
		{"sin plan nunca reclama", false, nil, false},
		{"sin plan aunque haga meses que no entrena", false, ago(90 * 24 * time.Hour), false},
		{"con plan y sin entrenar nunca", true, nil, true},
		{"con plan y entreno hoy", true, ago(2 * time.Hour), false},
		{"con plan y entreno hace 2 dias", true, ago(2 * 24 * time.Hour), false},
		{"justo en el umbral todavia no reclama", true, ago(AttentionThreshold), false},
		{"pasado el umbral reclama", true, ago(AttentionThreshold + time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsAttention(tt.hasAssignment, tt.lastSession, now); got != tt.want {
				t.Fatalf("NeedsAttention = %v, esperaba %v", got, tt.want)
			}
		})
	}
}
