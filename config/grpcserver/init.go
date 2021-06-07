package grpcserver

import (
	"context"
	api "github.com/trakkie-id/secondbaser/api/go_gen"
	"github.com/trakkie-id/secondbaser/config/application"
	"github.com/trakkie-id/secondbaser/service"
	"net"
	"os"
	"os/signal"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"google.golang.org/grpc"
)

func InitGrpcServer(grpcPort string) {
	listen, err := net.Listen("tcp", ":"+grpcPort)
	ctx := context.Background()

	if err != nil {
		panic(err)
	}

	//Init Dependency Injection
	trxService := service.TransactionServiceImpl()

	server := grpc.NewServer(
		grpc.StatsHandler(zipkingrpc.NewServerHandler(application.TRACER)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	api.RegisterTransactionalRequestServer(server, trxService)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			application.LOGGER.Info("Shutting down gRPC Server")

			server.Stop()

			<-ctx.Done()
		}
	}()

	// start gRPC server
	application.LOGGER.Info("gRPC Server Started, Listening on " + grpcPort)

	err = server.Serve(listen)

	if err != nil {
		panic(err)
	}
}
