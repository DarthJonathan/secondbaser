package application

import (
	"github.com/trakkie-id/secondbaser/config/logging"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apsdehal/go-logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
)

var (
	LOGGER          *logger.Logger
	AppName 		= "SECONDBASER"
	AppEnv			string
	TRACER          *zipkin.Tracer
	DB              *gorm.DB
	KafkaBroker		string
)

func SetUpLogger(logLevel string, AppName string) {
	LOGGER, _ = logger.New(AppName, 100, os.Stdout)
	LOGGER.SetFormat("%{time} [%{module}] [%{level}] %{message}")

	if strings.EqualFold(logLevel, "DEBUG") {
		LOGGER.SetLogLevel(logger.DebugLevel)
	} else {
		LOGGER.SetLogLevel(logger.InfoLevel)
	}
}

func InitAppEnv(env string) {
	if len(env) < 0 {
		AppEnv = "DEV"
	} else {
		AppEnv = env
	}
}

func InitDatabase(user string, password string, database string, host string, port string) {
	dsn := user + ":" + password + "@(" + host + ":" + port + ")/" + database + "?parseTime=true"

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

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 gormLogger,
	})

	if err != nil {
		panic(err)
	}

	//Config DB
	sqlDB, err := db.DB()

	if err != nil {
		panic(err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	//Set prometheus
	_ = db.Use(prometheus.New(prometheus.Config{
		DBName:          "trakkie_auth", // use `DBName` as metrics label
		RefreshInterval: 2000,           // Refresh metrics interval (default 15 seconds)
		StartServer:     false,          // start http server to expose metrics
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"threads_running"},
			},
		}, // user defined metrics
	}))

	LOGGER.Info("Database is connected")

	DB = db
}

func InitZipkinTracer(grpcPort string, endpointUrlConf string) {
	const defaultEndpointURL = "http://localhost:9411/api/v2/spans"

	if len(endpointUrlConf) < 1 {
		endpointUrlConf = defaultEndpointURL
	}

	reporter := zipkinhttpreporter.NewReporter(endpointUrlConf)

	localServer, err := zipkin.NewEndpoint(AppName, "localhost:"+grpcPort)

	if err != nil {
		LOGGER.Errorf("Error initializing zipkin! %s", err)
		panic(err)
	}

	sampler, err := zipkin.NewCountingSampler(1.0)
	if err != nil {
		LOGGER.Errorf("Error initializing zipkin! %s", err)
		panic(err)
	}

	t, err := zipkin.NewTracer(reporter, zipkin.WithSampler(sampler), zipkin.WithLocalEndpoint(localServer), zipkin.WithSharedSpans(false))

	if err != nil {
		LOGGER.Errorf("Error initializing zipkin! %s", err)
		panic(err)
	}

	LOGGER.Infof("Zipkin is connected, endpoint url : [%s]", endpointUrlConf)

	TRACER = t
}

func InitPrometheusServer(servingPort string) error {
	http.Handle("/prometheus", promhttp.Handler())
	LOGGER.Info("HTTP Server Started, listening on " + servingPort)
	return http.ListenAndServe(":"+servingPort, nil)
}