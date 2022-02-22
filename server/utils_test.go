package server

import (
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/marcuscarr/appts/models"
)

func TestBuildAvailable(t *testing.T) {
	appts := []models.Appt{
		{
			StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			EndTime:   time.Date(2020, 1, 1, 9, 30, 0, 0, location),
		},
		{
			StartTime: time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			EndTime:   time.Date(2020, 1, 1, 10, 30, 0, 0, location),
		},
		{
			StartTime: time.Date(2020, 1, 1, 11, 0, 0, 0, location),
			EndTime:   time.Date(2020, 1, 1, 11, 30, 0, 0, location),
		},
		{
			StartTime: time.Date(2020, 1, 1, 11, 30, 0, 0, location),
			EndTime:   time.Date(2020, 1, 1, 12, 0, 0, 0, location),
		},
	}

	expected := []time.Time{
		time.Date(2020, 1, 1, 9, 30, 0, 0, location),
		time.Date(2020, 1, 1, 10, 30, 0, 0, location),
		time.Date(2020, 1, 1, 12, 0o0, 0, 0, location),
		time.Date(2020, 1, 1, 12, 30, 0, 0, location),
	}

	available := buildAvailable(
		time.Date(2020, 1, 1, 9, 0, 0, 0, location),
		time.Date(2020, 1, 1, 13, 0, 0, 0, location),
		appts,
	)

	if len(available) != len(expected) {
		t.Errorf("Expected %d available times, got %d", len(expected), len(available))
	}

	for i, expectedTime := range expected {
		if available[i] != expectedTime {
			t.Errorf("Expected %v, got %v", expectedTime, available[i])
		}
	}
}

func TestIsUnavailable(t *testing.T) {
	unavailable := map[time.Time]struct{}{
		time.Date(2020, 1, 1, 9, 0, 0, 0, location):   {},
		time.Date(2020, 1, 1, 10, 0, 0, 0, location):  {},
		time.Date(2020, 1, 1, 11, 0, 0, 0, location):  {},
		time.Date(2020, 1, 1, 11, 30, 0, 0, location): {},
	}

	testCases := []struct {
		t time.Time
		e bool
	}{
		{time.Date(2020, 1, 1, 9, 0, 0, 0, location), true},
		{time.Date(2020, 1, 1, 9, 30, 0, 0, location), false},
		{time.Date(2020, 1, 1, 10, 0, 0, 0, location), true},
		{time.Date(2020, 1, 1, 10, 30, 0, 0, location), false},
		{time.Date(2020, 1, 1, 11, 0, 0, 0, location), true},
		{time.Date(2020, 1, 1, 11, 30, 0, 0, location), true},
	}

	for _, tc := range testCases {
		if isUnavailable(tc.t, unavailable) != tc.e {
			t.Errorf("Expected %v, got %v", tc.e, isUnavailable(tc.t, unavailable))
		}
	}
}

func TestHourMinuteBetween(t *testing.T) {
	testCases := []struct {
		start, end, check time.Time
		e                 bool
	}{
		{
			time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			time.Date(2020, 1, 1, 9, 30, 0, 0, location),
			true,
		},
		{
			time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			time.Date(2020, 1, 31, 9, 30, 0, 0, location),
			true,
		},
		{
			time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			false,
		},
		{
			time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 0, 0, 0, location),
			time.Date(2020, 1, 1, 10, 30, 0, 0, location),
			false,
		},
	}

	for _, tc := range testCases {
		if hourMinuteBetween(tc.start, tc.end, tc.check) != tc.e {
			t.Errorf("Expected %v, got %v", tc.e, hourMinuteBetween(tc.start, tc.end, tc.check))
		}
	}
}

func TestValidAppt(t *testing.T) {
	testCases := []struct {
		appt models.Appt
		e    error
	}{
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, location),
				EndTime:   time.Date(2020, 1, 1, 9, 30, 0, 0, location),
			},
			nil,
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, location),
				EndTime:   time.Date(2020, 1, 1, 10, 0o0, 0, 0, location),
			},
			errors.New("appt duration must be 30 minutes"),
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			},
			errors.New("Key: 'Appt.EndTime' Error:Field validation for 'EndTime' failed on the 'required' tag"),
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				EndTime:   time.Date(2020, 1, 1, 9, 0, 0, 0, location),
			},
			errors.New("Key: 'Appt.StartTime' Error:Field validation for 'StartTime' failed on the 'required' tag"),
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 5, 0, 0, 0, location),
				EndTime:   time.Date(2020, 1, 1, 5, 30, 0, 0, location),
			},
			errors.New("start time is outside of business hours"),
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 17, 0, 0, 0, location),
				EndTime:   time.Date(2020, 1, 1, 17, 30, 0, 0, location),
			},
			errors.New("start time is outside of business hours"),
		},
		{
			models.Appt{
				UserID:    1,
				TrainerID: 1,
				StartTime: time.Date(2020, 1, 1, 9, 15, 0, 0, location),
				EndTime:   time.Date(2020, 1, 1, 9, 45, 0, 0, location),
			},
			errors.New("appt time must be on the hour a multiple of 30 minutes past the hour"),
		},
	}

	validator := validator.New()
	for _, tc := range testCases {
		if err := validAppt(validator, tc.appt); err != tc.e {
			if err.Error() != tc.e.Error() {
				t.Errorf("Expected %v, got %v", tc.e, err)
			}
		}
	}
}
