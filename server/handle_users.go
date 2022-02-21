package server

import (
	"gorm.io/gorm"

	"github.com/marcuscarr/appts/models"
)

type userHandler struct {
	*modelHandler
}

func newUserHandler(db *gorm.DB) *userHandler {
	return &userHandler{
		modelHandler: newModelHandler(db, &models.User{}, "id", nil),
	}
}
