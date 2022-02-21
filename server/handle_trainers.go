package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gorilla/mux"
	"github.com/marcuscarr/appts/models"
)

const dateFormat = "2006-01-02"

type trainerHandler struct {
	*modelHandler
}

func newTrainerHandler(db *gorm.DB) *trainerHandler {
	return &trainerHandler{
		modelHandler: newModelHandler(db, &models.Trainer{}, "id", nil),
	}
}

func (th *trainerHandler) getAvailableAppts(w http.ResponseWriter, r *http.Request) {
	var wheres []string
	var args []interface{}

	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse trainer_id
	vars := mux.Vars(r)
	trainerID, err := strconv.Atoi(vars[trainerIDParam])
	if err != nil {
		log.Printf("Error parsing trainer_id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse query params
	query := r.URL.Query()

	// Parse start_date
	if len(query[startsAtParam]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	starteDateVal := query[startsAtParam][0]
	startDatetime, err := time.Parse(dateFormat, starteDateVal)
	if err != nil {
		log.Printf("Error parsing starts_at: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse end_date
	if len(query[endsAtParam]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endDateVal := query[endsAtParam][0]
	endDatetime, err := time.Parse(dateFormat, endDateVal)
	if err != nil {
		log.Printf("Error parsing ends_at: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	wheres = append(wheres, "trainer_id = ?")
	args = append(args, trainerID)

	// We want to find all appts that start before the end date and end after the start date, so
	// set the start date to the start of the day and the end date to the end of the day.
	startDate := time.Date(startDatetime.Year(), startDatetime.Month(), startDatetime.Day(), 0, 0, 0, 0, location)
	wheres = append(wheres, "start_time >= ?")
	args = append(args, startDate)

	endDate := time.Date(endDatetime.Year(), endDatetime.Month(), endDatetime.Day(), 23, 59, 59, 0, location)
	wheres = append(wheres, "end_time <= ?")
	args = append(args, endDate)

	var appts []models.Appt
	where := strings.Join(wheres, " AND ")
	dbErr := th.db.Where(where, args...).Find(&appts)
	if dbErr.Error != nil {
		log.Printf("Error finding appts: %v", dbErr.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res []string
	for _, a := range buildAvailable(startDate, endDate, appts) {
		res = append(res, a.Format(time.RFC3339))
	}

	err = json.NewEncoder(w).Encode(map[string][]string{"available": res})
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
