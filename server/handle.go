package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type modelHandler struct {
	db *gorm.DB

	model reflect.Type

	idParam string
	queries []queries
}

type queries struct {
	param string
	op    string
}

func newModelHandler(db *gorm.DB, model interface{}, idParam string, queries []queries) *modelHandler {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Ptr {
		panic("model must be a pointer")
	}

	modelValue := reflect.ValueOf(model)
	if modelValue.IsNil() {
		panic("model must not be nil")
	}

	modelType = modelValue.Elem().Type()
	if modelType.Kind() != reflect.Struct {
		panic("model must be a struct")
	}

	return &modelHandler{
		db:      db,
		model:   modelType,
		idParam: idParam,
		queries: queries,
	}
}

func (mh *modelHandler) create(w http.ResponseWriter, r *http.Request) {
	model := reflect.New(mh.model).Interface()

	if err := json.NewDecoder(r.Body).Decode(model); err != nil {
		log.Printf("Error decoding body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If the passed model has an ID, update the auto-increment sequence.
	id := reflect.ValueOf(model).Elem().FieldByName("ID")
	if id.IsValid() && id.Uint() != 0 {
		err := mh.createWithID(mh.db, model)
		if err != nil {
			log.Printf("Error creating model with ID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		if result := mh.db.Create(model); result.Error != nil {
			log.Printf("Error creating model: %v", result.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err := json.NewEncoder(w).Encode(model)
	if err != nil {
		log.Printf("Error encoding model: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (mh *modelHandler) createWithID(db *gorm.DB, model interface{}) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(model); result.Error != nil {
			return result.Error
		}

		// Gorm default table name is the model name in lowercase plural.
		tableName := strings.ToLower(mh.model.Name()) + "s"
		log.Printf("Setting sequence for table name: %s", tableName)
		result := tx.Exec("SELECT setval(pg_get_serial_sequence(?, 'id'), max(id)) FROM "+tableName, tableName)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})
}

func (mh *modelHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[mh.idParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	model := reflect.New(mh.model).Interface()
	result := mh.db.First(model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(model)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *modelHandler) list(w http.ResponseWriter, r *http.Request) {
	var wheres []interface{}
	var args []interface{}
	for _, q := range mh.queries {
		value := r.URL.Query().Get(q.param)
		if value == "" {
			continue
		}

		wheres = append(wheres, fmt.Sprintf("%s %s ?", q.param, q.op))
		args = append(args, value)
	}

	models := reflect.New(reflect.SliceOf(mh.model)).Interface()

	if len(wheres) > 0 {
		result := mh.db.Where(wheres, args...).Find(models)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		result := mh.db.Find(models)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err := json.NewEncoder(w).Encode(models)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (mh *modelHandler) update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[mh.idParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	modelValue := reflect.New(mh.model)
	modelValue.FieldByName("ID").SetInt(int64(id))
	model := modelValue.Interface()

	if err := json.NewDecoder(r.Body).Decode(model); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if result := mh.db.Save(model); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(model)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (mh *modelHandler) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idParam := vars[mh.idParam]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	modelValue := reflect.New(mh.model)
	modelValue.FieldByName("ID").SetInt(int64(id))
	model := modelValue.Interface()

	if result := mh.db.Delete(model); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
