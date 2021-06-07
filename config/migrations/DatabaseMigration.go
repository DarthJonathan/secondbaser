package migrations

import (
	"github.com/trakkie-id/secondbaser/config/application"
	"github.com/trakkie-id/secondbaser/model"
)

func MigrateDatabase() {
	_ = application.DB.AutoMigrate(&model.Transaction{}, &model.TransactionParticipant{})
}