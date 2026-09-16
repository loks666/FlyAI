package service

import (
	"testing"
	"time"
)

func TestValidateFeatureTime(t *testing.T) {
	for _, value := range []string{"00:00", "08:00", "23:59"} {
		if err := ValidateFeatureTime(value); err != nil {
			t.Fatalf("valid time %s rejected: %v", value, err)
		}
	}
	for _, value := range []string{"8:00", "24:00", "08:60", "abcde"} {
		if err := ValidateFeatureTime(value); err == nil {
			t.Fatalf("invalid time %s accepted", value)
		}
	}
}

func TestFeatureSchedule(t *testing.T) {
	monday := time.Date(2026, 8, 24, 9, 30, 0, 0, time.Local)
	tuesday := monday.AddDate(0, 0, 1)
	for _, schedule := range []string{ScheduleDaily, ScheduleMonday, ScheduleSunday, ScheduleMonthly, ScheduleYearly} {
		if err := ValidateFeatureSchedule(schedule); err != nil {
			t.Fatalf("valid schedule %s rejected: %v", schedule, err)
		}
	}
	if err := ValidateFeatureSchedule("hourly"); err == nil {
		t.Fatal("invalid schedule accepted")
	}
	if !FeatureScheduleMatches(ScheduleDaily, monday) || !FeatureScheduleMatches(ScheduleDaily, tuesday) {
		t.Fatal("daily schedule should match every day")
	}
	if !FeatureScheduleMatches(ScheduleMonday, monday) || FeatureScheduleMatches(ScheduleMonday, tuesday) {
		t.Fatal("Monday schedule matched the wrong weekday")
	}
	monthStart := time.Date(2026, 9, 1, 9, 30, 0, 0, time.Local)
	if !FeatureScheduleMatches(ScheduleMonthly, monthStart) || FeatureScheduleMatches(ScheduleMonthly, monthStart.AddDate(0, 0, 1)) {
		t.Fatal("monthly schedule must only match the first day of a month")
	}
	yearStart := time.Date(2027, 1, 1, 9, 0, 0, 0, time.Local)
	if !FeatureScheduleMatches(ScheduleYearly, yearStart) || FeatureScheduleMatches(ScheduleYearly, monthStart) {
		t.Fatal("yearly schedule must only match January 1")
	}
}
