package server

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/marcuscarr/appts/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	// Params
	apptIDParam    = "appt_id"
	trainerIDParam = "trainer_id"
	userIDParam    = "user_id"
	startDateParam = "start_date"
	endDateParam   = "end_date"
)

type Server struct {
	db     *gorm.DB
	router *mux.Router
}

func (s *Server) routes() {
	apptsRouter := s.router.PathPrefix("/appointments").Subrouter()

	apptsRouter.HandleFunc("", s.createAppt).Methods("POST")
	apptsRouter.HandleFunc("", s.getAppts).Methods("GET")

	apptIDRoute := fmt.Sprintf("/{%s}", apptIDParam)
	apptsRouter.HandleFunc(apptIDRoute, s.getAppt).Methods("GET")
	apptsRouter.HandleFunc(apptIDRoute, s.updateAppt).Methods("PUT")
	apptsRouter.HandleFunc(apptIDRoute, s.deleteAppt).Methods("DELETE")

	trainersRouter := s.router.PathPrefix("/trainers").Subrouter()

	trainersRouter.HandleFunc("", s.createTrainer).Methods("POST")
	trainersRouter.HandleFunc("", s.getTrainers).Methods("GET")

	trainerIDRoute := fmt.Sprintf("/{%s}", trainerIDParam)
	trainersRouter.HandleFunc(trainerIDRoute, s.getTrainer).Methods("GET")
	trainersRouter.HandleFunc(trainerIDRoute, s.updateTrainer).Methods("PUT")
	trainersRouter.HandleFunc(trainerIDRoute, s.deleteTrainer).Methods("DELETE")

	trainerApptsRouter := trainersRouter.PathPrefix("/appointments").Subrouter()
	trainerApptsRouter.HandleFunc("", s.getAppts).Methods("GET")
	trainerApptsRouter.HandleFunc("/available", s.getAvailableAppts).
		Methods("GET").
		Queries(
			startDateParam, fmt.Sprintf("{%s}", startDateParam),
			endDateParam, fmt.Sprintf("{%s}", endDateParam),
		)

	usersRouter := s.router.PathPrefix("/users").Subrouter()

	usersRouter.HandleFunc("", s.createUser).Methods("POST")
	usersRouter.HandleFunc("", s.getUsers).Methods("GET")

	userIDRoute := fmt.Sprintf("/{%s}", userIDParam)
	usersRouter.HandleFunc(userIDRoute, s.getUser).Methods("GET")
	usersRouter.HandleFunc(userIDRoute, s.updateUser).Methods("PUT")
	usersRouter.HandleFunc(userIDRoute, s.deleteUser).Methods("DELETE")

	usersRouter.HandleFunc(userIDRoute+"/appointments", s.getAppts).Methods("GET")
}

func newServer() *Server {
	db, err := gorm.Open(sqlite.Open("appointments.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&models.Appt{}, &models.User{}, &models.Trainer{})
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	s := &Server{
		db:     db,
		router: router,
	}

	s.routes()

	return &Server{
		db:     db,
		router: mux.NewRouter(),
	}
}
