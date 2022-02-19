package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/marcuscarr/appts/models"
)

const (
	apptDuration = 30 * time.Minute
	locale       = "America/Los_Angeles"

	dateFormat = "2006-01-02"
	timeFormat = "15:04:05-07:00"
)

var (
	location, _ = time.LoadLocation(locale)
	startOfDay  = time.Date(0, 0, 0, 8, 0, 0, 0, location)
	endOfDay    = time.Date(0, 0, 0, 17, 0, 0, 0, location)
)

func (s *Server) createAppt(w http.ResponseWriter, r *http.Request) {
	// TODO: Add validation of time, duration and availaility.

	var appt models.Appt
	if err := json.NewDecoder(r.Body).Decode(&appt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.db.Create(&appt)
	err := json.NewEncoder(w).Encode(appt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getAppts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var wheres []string
	var args []interface{}

	if trainerID := query.Get(trainerIDParam); trainerID != "" {
		parsed, err := strconv.Atoi(trainerID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wheres = append(wheres, "trainer_id = ?")
		args = append(args, parsed)
	}

	if userID := query.Get(userIDParam); userID != "" {
		parsed, err := strconv.Atoi(userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wheres = append(wheres, "user_id = ?")
		args = append(args, parsed)
	}

	var appts []models.Appt

	if len(wheres) > 0 {
		where := strings.Join(wheres, " AND ")
		s.db.Where(where, args...).Find(&appts)
	} else {
		s.db.Find(&appts)
	}

	err := json.NewEncoder(w).Encode(appts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getAppt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[apptIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var appt models.Appt
	s.db.First(&appt, id)

	if appt.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(appt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateAppt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[apptIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existingAppt models.Appt
	s.db.First(&existingAppt, id)

	if existingAppt.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var appt models.Appt
	if err := json.NewDecoder(r.Body).Decode(&appt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	appt.ID = uint(id)

	s.db.Model(&appt).Updates(appt)
	err = json.NewEncoder(w).Encode(appt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteAppt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[apptIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var appt models.Appt
	s.db.First(&appt, id)

	if appt.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.db.Delete(&appt)
}

func (s *Server) getAvailableAppts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var wheres []string
	var args []interface{}

	if trainerID := query.Get(trainerIDParam); trainerID != "" {
		parsed, err := strconv.Atoi(trainerID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wheres = append(wheres, "trainer_id = ?")
		args = append(args, parsed)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var start time.Time
	if startDate := query.Get(startDateParam); startDate != "" {
		parsed, err := time.Parse(dateFormat, startDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		wheres = append(wheres, "start_time >= ?")
		args = append(args, parsed)
		start = parsed
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var end time.Time
	if endDate := query.Get(endDateParam); endDate != "" {
		parsedDate, err := time.Parse(dateFormat, endDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Add one day to make the range inclusive.
		parsed := parsedDate.AddDate(0, 0, 1)

		wheres = append(wheres, "end_time < ?")
		args = append(args, parsed)
		end = parsed
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	where := strings.Join(wheres, " AND ")

	var appts []models.Appt
	s.db.Where(where, args...).Find(&appts)

	unavailable := make(map[time.Time]struct{})
	for _, appt := range appts {
		unavailable[appt.StartTime] = struct{}{}
	}

	var available []time.Time
	nextAppt := time.Date(
		start.Year(), start.Month(), start.Day(),
		startOfDay.Hour(), startOfDay.Minute(), startOfDay.Second(), 0, location,
	)
	for nextAppt.Before(end) {
		if _, ok := unavailable[nextAppt]; !ok && hourMinuteBetween(startOfDay, endOfDay, nextAppt) {
			available = append(available, nextAppt)
		}
		nextAppt = nextAppt.Add(apptDuration)
	}
	var res []string
	for _, a := range available {
		res = append(res, a.Format(timeFormat))
	}

	err := json.NewEncoder(w).Encode(map[string][]string{"available": res})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func buildAvailable(start, end time.Time, appts []models.Appt) []time.Time {
	unavailable := make(map[time.Time]struct{})
	for _, appt := range appts {
		unavailable[appt.StartTime] = struct{}{}
	}

	var available []time.Time
	nextAppt := time.Date(
		start.Year(), start.Month(), start.Day(),
		startOfDay.Hour(), startOfDay.Minute(), startOfDay.Second(), 0, location,
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

	if end.Hour() == check.Hour() && end.Minute() >= check.Minute() {
		return true
	}

	return false
}
