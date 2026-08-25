package types

import (
	"testing"
	"time"
)

func TestDateTimeScan(t *testing.T) {
	want := time.Date(2026, 8, 21, 15, 4, 5, 123000000, time.Local)
	tests := []struct {
		name  string
		value any
	}{
		{name: "time", value: want},
		{name: "bytes", value: []byte("2026-08-21 15:04:05.123")},
		{name: "string", value: "2026-08-21 15:04:05.123"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got DateTime
			if err := got.Scan(test.value); err != nil {
				t.Fatalf("Scan() error = %v", err)
			}
			if !got.ToTime().Equal(want) {
				t.Fatalf("Scan() = %v, want %v", got.ToTime(), want)
			}
		})
	}
}

func TestDateTimeScanNil(t *testing.T) {
	var got DateTime
	if err := got.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if !got.IsNil() {
		t.Fatalf("Scan(nil) should produce zero DateTime, got %v", got)
	}
}

func TestDateTimeScanRejectsUnknownValue(t *testing.T) {
	var got DateTime
	if err := got.Scan(42); err == nil {
		t.Fatal("Scan(42) should return an error")
	}
}
