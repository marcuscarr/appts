package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/marcuscarr/appts/models"
)

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.db.Create(&user)
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	s.db.Find(&users)
	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[userIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user models.User
	s.db.First(&user, id)

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[userIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var existingUser models.User
	s.db.First(&existingUser, id)

	if existingUser.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.ID = uint(id)

	s.db.Model(&user).Updates(user)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[userIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user models.User
	s.db.First(&user, id)

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.db.Delete(&user)
}

func (s *Server) getUserAppts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[userIDParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user models.User
	s.db.Preload("Appts").First(&user, id)

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	appts := user.Appts
	err json.NewEncoder(w).Encode(appts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
