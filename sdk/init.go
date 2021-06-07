package sdk

import (
	"database/sql"
	"github.com/apsdehal/go-logger"
	"github.com/openzipkin/zipkin-go"
	"github.com/trakkie-id/secondbaser/config/logging"
	"github.com/trakkie-id/secondbaser/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

var (
	LOGGER          *logger.Logger
	AppName 		string
	TRACER          *zipkin.Tracer
	DB              *gorm.DB
	KafkaAddress 	string
	KafkaGroupId	string
	Server string
)

func InitSecondbaserClient(db *sql.DB, tracer *zipkin.Tracer, logLevel string, kafkaBrokerAddr string, appGroup string, appName string, secondBaserServer string) {
	//Zipkin
	TRACER = tracer

	//Setup Logger
	setUpLogger(logLevel)

	//Setup Database
	setupDatabase(db)

	//Setup Kafka
	KafkaAddress = kafkaBrokerAddr
	KafkaGroupId = appGroup

	//Setup app name
	AppName = appName

	//Second Baser Server
	Server = secondBaserServer

	//Migrate tables
	_ = DB.AutoMigrate(&model.Transaction{}, &model.TransactionParticipant{})
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
	LOGGER, _ = logger.New("SECONDBASER-CLIENT", 100, os.Stdout)
	LOGGER.SetFormat("%{time} [%{module}] [%{level}] %{message}")

	if strings.EqualFold(logLevel, "DEBUG") {
		LOGGER.SetLogLevel(logger.DebugLevel)
	} else {
		LOGGER.SetLogLevel(logger.InfoLevel)
	}
}