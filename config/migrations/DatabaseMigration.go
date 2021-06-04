package migrations

import (
	"github.com/DarthJonathan/secondbaser/config/application"
	"github.com/DarthJonathan/secondbaser/model"
)

func MigrateDatabase() {
	_ = application.DB.AutoMigrate(&model.Transaction{}, &model.TransactionParticipant{})
}