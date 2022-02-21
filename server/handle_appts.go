package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/marcuscarr/appts/models"
)

const (
	apptDuration = 30 * time.Minute
	locale       = "America/Los_Angeles"
)

var (
	location, _ = time.LoadLocation(locale)
	startOfDay  = time.Date(0, 0, 0, 8, 0, 0, 0, location)
	endOfDay    = time.Date(0, 0, 0, 17, 0, 0, 0, location)
	// use a single instance of Validate, it caches struct info
)

type apptHandler struct {
	*modelHandler
	validator *validator.Validate
}

func newApptHandler(db *gorm.DB) *apptHandler {
	validate := validator.New()
	return &apptHandler{
		modelHandler: newModelHandler(
			db, &models.Appt{}, "id",
			[]queries{
				{"user_id", "="},
				{"trainer_id", "="},
				{"start_time", ">="},
				{"end_time", "<"},
			},
		),
		validator: validate,
	}
}

func (ah *apptHandler) create(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate request.
	var appt models.Appt

	if err := json.NewDecoder(r.Body).Decode(&appt); err != nil {
		log.Printf("Error decoding body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validAppt(ah.validator, appt); err != nil {
		log.Printf("Invalid appt: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	txErr := ah.db.Transaction(func(tx *gorm.DB) error {
		isAvailable, err := availableAppt(tx, appt)
		if err != nil {
			log.Printf("Error checking if appt is available: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if !isAvailable {
			log.Printf("Appt is not available")
			w.WriteHeader(http.StatusConflict)
			return errors.New("appt is not available")
		}

		if appt.ID != 0 {
			var apptValue interface{} = &appt
			result := ah.modelHandler.createWithID(tx, apptValue)
			if result != nil {
				log.Printf("Error creating appt: %v", result)
				w.WriteHeader(http.StatusInternalServerError)
				return result
			}
		} else {
			result := tx.Create(&appt)
			if result.Error != nil {
				log.Printf("Error creating appt: %v", result.Error)
				w.WriteHeader(http.StatusInternalServerError)
				return result.Error
			}
		}

		return nil
	})
	if txErr != nil {
		w.Write([]byte(txErr.Error()))
		return
	}

	err := json.NewEncoder(w).Encode(appt)
	if err != nil {
		log.Printf("Error encoding appt: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ah *apptHandler) updateAppt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idValue := vars[apptIDParam]
	id, err := strconv.Atoi(idValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existingAppt models.Appt
	if result := ah.db.First(&existingAppt, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ah.create(w, r)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var appt models.Appt
	if err := json.NewDecoder(r.Body).Decode(&appt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	appt.ID = uint(id)

	if err := validAppt(ah.validator, appt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	txErr := ah.db.Transaction(func(tx *gorm.DB) error {
		isAvailable, err := availableAppt(tx, appt)
		if err != nil {
			log.Printf("Error checking if appt is available: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		if !isAvailable {
			log.Printf("Appt is not available")
			w.WriteHeader(http.StatusConflict)
			return errors.New("appt is not available")
		}

		result := tx.Save(&appt)
		if result.Error != nil {
			log.Printf("Error updating appt: %v", result.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return result.Error
		}

		return nil
	})

	if txErr != nil {
		w.Write([]byte(txErr.Error()))
		return
	}

	err = json.NewEncoder(w).Encode(appt)
	if err != nil {
		log.Printf("Error encoding appt: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func availableAppt(db *gorm.DB, appt models.Appt) (bool, error) {
	var existing []models.Appt
	result := db.Where("trainer_id = ? AND start_time = ?", appt.TrainerID, appt.StartTime).First(&existing)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, result.Error
	}

	return len(existing) == 0, nil
}
