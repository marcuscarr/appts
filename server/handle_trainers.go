package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/marcuscarr/appts/models"
)

func (s *Server) createTrainer(w http.ResponseWriter, r *http.Request) {
	var trainer models.Trainer
	if err := json.NewDecoder(r.Body).Decode(&trainer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.db.Create(&trainer)
	err = json.NewEncoder(w).Encode(trainer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getTrainers(w http.ResponseWriter, r *http.Request) {
	var trainers []models.Trainer
	s.db.Find(&trainers)
	err := json.NewEncoder(w).Encode(trainers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getTrainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[trainerIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var trainer models.Trainer
	s.db.First(&trainer, id)

	if trainer.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(trainer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateTrainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[trainerIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existingTrainer models.Trainer
	s.db.First(&existingTrainer, id)

	if existingTrainer.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var trainer models.Trainer
	if err := json.NewDecoder(r.Body).Decode(&trainer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.db.Model(&existingTrainer).Updates(trainer)
	err = json.NewEncoder(w).Encode(existingTrainer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteTrainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[trainerIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existingTrainer models.Trainer
	s.db.First(&existingTrainer, id)

	if existingTrainer.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.db.Delete(&existingTrainer)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getTrainerAppts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[trainerIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var trainer models.Trainer
	s.db.Preload("Appts").First(&trainer, id)

	if trainer.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	appts := trainer.Appts
	err = json.NewEncoder(w).Encode(appts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
