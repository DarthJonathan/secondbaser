package sdk

import (
	"database/sql"
	"github.com/DarthJonathan/secondbaser/config/logging"
	"github.com/apsdehal/go-logger"
	"github.com/openzipkin/zipkin-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

var (
	LOGGER          *logger.Logger
	AppName 		= "SECONDBASER-CLIENT"
	TRACER          *zipkin.Tracer
	DB              *gorm.DB
)

func InitializeSDK(db *sql.DB, tracer *zipkin.Tracer, logLevel string) {
	//Zipkin
	TRACER = tracer

	//Setup Logger
	setUpLogger(logLevel)

	//Setup Database
	setupDatabase(db)

	//Setup Kafka

}

func setupDatabase(db *sql.DB) {
	gormLogger := gorm_logger.New(
		&logging.GormLogger{
			Logger: LOGGER,
		}, // io writer
		gorm_logger.Config{
			SlowThreshold: time.Second,      // Slow SQL threshold
			LogLevel:      gorm_logger.Info, // Log level
			Colorful:      true,             // Disable color
		},
	)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 gormLogger,
	})

	if err != nil {
		LOGGER.Errorf("Cannot connect to database!")
	}

	DB = gormDB
}

func setUpLogger(logLevel string) {
	LOGGER, _ = logger.New(AppName, 100, os.Stdout)
	LOGGER.SetFormat("%{time} [%{module}] [%{level}] %{message}")

	if strings.EqualFold(logLevel, "DEBUG") {
		LOGGER.SetLogLevel(logger.DebugLevel)
	} else {
		LOGGER.SetLogLevel(logger.InfoLevel)
	}
}

func setupKafka() {

}
