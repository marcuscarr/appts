package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/marcuscarr/appts/models"
)

func buildAvailable(start, end time.Time, appts []models.Appt) []time.Time {
	unavailable := make(map[time.Time]struct{})
	for _, appt := range appts {
		unavailable[appt.StartTime] = struct{}{}
	}

	var available []time.Time
	nextAppt := time.Date(
		start.Year(), start.Month(), start.Day(),
		start.Hour(), start.Minute(), start.Second(), 0, location,
	)
	for nextAppt.Before(end) {
		if !isUnavailable(nextAppt, unavailable) && hourMinuteBetween(startOfDay, endOfDay, nextAppt) {
			available = append(available, nextAppt)
		}
		nextAppt = nextAppt.Add(apptDuration)
	}

	return available
}

func isUnavailable(t time.Time, unavailable map[time.Time]struct{}) bool {
	_, ok := unavailable[t]
	return ok
}

func hourMinuteBetween(start, end, check time.Time) bool {
	if start.Hour() < check.Hour() && end.Hour() > check.Hour() {
		return true
	}

	if start.Hour() == check.Hour() && start.Minute() <= check.Minute() {
		return true
	}

	if end.Hour() == check.Hour() && end.Minute() > check.Minute() {
		return true
	}

	return false
}

func validAppt(validator *validator.Validate, appt models.Appt) error {
	if err := validator.Struct(appt); err != nil {
		return err
	}

	startTime := time.Date(
		startOfDay.Year(), startOfDay.Month(), startOfDay.Day(),
		appt.StartTime.Hour(), appt.StartTime.Minute(), 0, 0, location,
	)
	if startTime.Before(startOfDay) || startTime == endOfDay || startTime.After(endOfDay) {
		return errors.New("start time is outside of business hours")
	}

	endTime := time.Date(
		endOfDay.Year(), endOfDay.Month(), endOfDay.Day(),
		appt.EndTime.Hour(), appt.EndTime.Minute(), 0, 0, location,
	)
	if endTime.Before(startOfDay) || endTime.After(endOfDay) {
		return errors.New("end time is outside of business hours")
	}

	if appt.EndTime.Sub(appt.StartTime) != apptDuration {
		return fmt.Errorf("appt duration must be %.0f minutes", apptDuration.Minutes())
	}

	if (appt.StartTime.Minute())%int(apptDuration.Minutes()) != 0 {
		return fmt.Errorf(
			"appt time must be on the hour a multiple of %.0f minutes past the hour",
			apptDuration.Minutes(),
		)
	}

	return nil
}
