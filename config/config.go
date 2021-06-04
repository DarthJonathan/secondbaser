package config

import (
	"flag"
	"github.com/DarthJonathan/secondbaser/config/application"
	"github.com/DarthJonathan/secondbaser/config/grpcserver"
	"github.com/DarthJonathan/secondbaser/config/migrations"

	"github.com/tkanos/gonfig"
	"golang.org/x/sync/errgroup"
)

func RunServer() error {
	var configFile string
	var cfg Config

	flag.StringVar(&configFile, "config-file", "./config/conf/development.json", "Application configuration file")
	flag.Parse()

	err := gonfig.GetConf(configFile, &cfg)

	//Setup Logger
	application.SetUpLogger(cfg.LogLevel, application.AppName)

	//Print config info
	application.LOGGER.DebugF("Loaded application configuration, application configuration : %s", cfg)

	if err != nil {
		panic(err)
	}

	//Init App Env
	application.InitAppEnv(cfg.ApplicationEnv)

	//Init Database
	application.InitDatabase(cfg.DBUser, cfg.DBPassword, cfg.DBDatabase, cfg.DBHost, cfg.DBPort)

	//Migrate Database
	migrations.MigrateDatabase()

	//Init Tracer
	application.InitZipkinTracer(cfg.GRPCPort, cfg.ZipkinEndpoint)

	//Init Kafka
	application.KafkaBroker = cfg.KafkaBrokerAddress

	//Use Error Group for Threads
	g := new(errgroup.Group)

	//Init Prometheus Endpoint
	g.Go(func() error {
		return application.InitPrometheusServer(cfg.HTTPPort)
	})

	//Init Grpc Server blocking
	grpcserver.InitGrpcServer(cfg.GRPCPort)

	return nil
}
