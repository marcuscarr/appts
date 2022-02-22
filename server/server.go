package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/marcuscarr/appts/models"
)

const (
	// Params
	idParam        = "id"
	apptIDParam    = "appt_id"
	trainerIDParam = "trainer_id"
	userIDParam    = "user_id"
	startsAtParam  = "starts_at"
	endsAtParam    = "ends_at"
	startTimeParam = "start_time"
	endTimeParam   = "end_time"
)

type Server struct {
	http.Server
	db      *gorm.DB
	router  *mux.Router
	closers []io.Closer
	config  *Config
}

type Config struct {
	Host    string
	Port    int
	DBHost  string
	DBPort  int
	DBUser  string
	DBPass  string
	DBName  string
	Timeout time.Duration
}

func (s *Server) routes() {
	s.router.HandleFunc("/healthz", s.healthz).Methods("GET")

	apptHandler := newApptHandler(s.db)
	apptsRouter := s.router.PathPrefix("/appointments").Subrouter()

	apptsRouter.HandleFunc("", apptHandler.create).Methods("POST")
	apptsRouter.HandleFunc("", apptHandler.list).Methods("GET")

	apptIDRoute := fmt.Sprintf("/{%s}", idParam)
	apptsRouter.HandleFunc(apptIDRoute, apptHandler.get).Methods("GET")
	apptsRouter.HandleFunc(apptIDRoute, apptHandler.update).Methods("PUT")
	apptsRouter.HandleFunc(apptIDRoute, apptHandler.delete).Methods("DELETE")

	trainerHandler := newTrainerHandler(s.db)
	trainersRouter := s.router.PathPrefix("/trainers").Subrouter()

	trainersRouter.HandleFunc("", trainerHandler.create).Methods("POST")
	trainersRouter.HandleFunc("", trainerHandler.get).Methods("GET")

	trainerIDRoute := fmt.Sprintf("/{%s}", idParam)
	trainersRouter.HandleFunc(trainerIDRoute, trainerHandler.list).Methods("GET")
	trainersRouter.HandleFunc(trainerIDRoute, trainerHandler.update).Methods("PUT")
	trainersRouter.HandleFunc(trainerIDRoute, trainerHandler.delete).Methods("DELETE")

	trainerApptsRoute := fmt.Sprintf("/{%s}/appointments", trainerIDParam)
	trainersRouter.HandleFunc(trainerApptsRoute, apptHandler.list).Methods("GET")
	trainersRouter.HandleFunc(trainerApptsRoute+"/available", trainerHandler.getAvailableAppts).
		Methods("GET").
		Queries(
			startsAtParam, fmt.Sprintf("{%s}", startsAtParam),
			endsAtParam, fmt.Sprintf("{%s}", endsAtParam),
		)

	userHandler := newUserHandler(s.db)
	usersRouter := s.router.PathPrefix("/users").Subrouter()

	usersRouter.HandleFunc("", userHandler.create).Methods("POST")
	usersRouter.HandleFunc("", userHandler.list).Methods("GET")

	userIDRoute := fmt.Sprintf("/{%s}", idParam)
	usersRouter.HandleFunc(userIDRoute, userHandler.get).Methods("GET")
	usersRouter.HandleFunc(userIDRoute, userHandler.update).Methods("PUT")
	usersRouter.HandleFunc(userIDRoute, userHandler.delete).Methods("DELETE")

	usersRouter.HandleFunc(userIDRoute+"/appointments", apptHandler.list).Methods("GET")
}

func New(config *Config) *Server {
	// Connect to the postgres instance
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		config.DBHost, strconv.Itoa(config.DBPort), config.DBUser, config.DBPass,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Create the database if it doesn't exist
	db = db.Exec(fmt.Sprintf(`
		SELECT 'CREATE DATABASE %s'
		WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '%s');
	`, config.DBName, config.DBName))
	if db.Error != nil {
		log.Fatal(db.Error)
	}

	dbConn, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	dbConn.Close()

	// Connect to the database
	dsn = fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		config.DBHost, strconv.Itoa(config.DBPort), config.DBUser, config.DBName, config.DBPass,
	)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Trainer{}, &models.Appt{})
	if err != nil {
		panic(err)
	}

	// Get the database connection for closing
	dbConn, err = db.DB()
	if err != nil {
		panic(err)
	}

	// Create the server
	r := mux.NewRouter()
	s := &Server{
		Server: http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
			IdleTimeout:  60 * time.Second,
			Handler:      r,
		},
		db:      db,
		closers: []io.Closer{dbConn},
		router:  r,
	}
	s.routes()

	return s
}

func (s *Server) Run() {
	defer close(s.closers...)

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Starting server on %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C).
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	err := s.Shutdown(ctx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("shutting down")
	os.Exit(0)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func close(closers ...io.Closer) {
	var wg sync.WaitGroup
	wg.Add(len(closers))
	for _, c := range closers {
		go func(c io.Closer) {
			defer wg.Done()
			if err := c.Close(); err != nil {
				log.Println(err)
			}
		}(c)
	}

	wg.Wait()
}
