package sdk

import (
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"google.golang.org/grpc"
)

func GetConn() (*grpc.ClientConn,error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(Server, grpc.WithInsecure(), grpc.WithStatsHandler(zipkingrpc.NewClientHandler(TRACER)))
	if err != nil {
		return nil,err
	}
	return conn, nil
}

func CloseConn(conn *grpc.ClientConn) {
	defer conn.Close()
}
