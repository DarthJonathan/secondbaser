package config

import (
	"flag"

	"github.com/tkanos/gonfig"
	"github.com/trakkie-id/trakkie-member/config/conf"
	"github.com/trakkie-id/trakkie-member/config/grpcserver"
	"golang.org/x/sync/errgroup"
)

func RunServer() error {
	var configFile string
	var cfg conf.Config

	flag.StringVar(&configFile, "config-file", "./config/conf/development.json", "Application configuration file")
	flag.Parse()

	err := gonfig.GetConf(configFile, &cfg)

	//Setup Logger
	application.SetUpLogger(cfg.LogLevel, application.AppName)

	//Override Config Info
	application.OverrideEnvVars(&cfg)

	//Print config info
	application.LOGGER.DebugF("Loaded application configuration file, application configuration : %s", cfg)

	if err != nil {
		panic(err)
	}

	//Init App Env
	application.InitAppEnv(cfg.ApplicationEnv)

	//Init client address
	application.InitClients(cfg.ClientAddress)

	//Init Database
	application.InitDatabase(cfg.DBUser, cfg.DBPassword, cfg.DBDatabase, cfg.DBHost, cfg.DBPort)

	//Init Tracer
	application.InitZipkinTracer(cfg.GRPCPort, cfg.ZipkinEndpoint)

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
