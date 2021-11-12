package grpcserver

import (
	. "github.com/jinuopti/lilpop-server/log"
	"google.golang.org/grpc"
	"net"
	example "github.com/jinuopti/lilpop-server/communication/grpc/example"
)

func GrpcServer(port string) {
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		Loge("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	example.RegisterUserServer(grpcServer, &example.UserServ{})

	Logd("start gRPC server on %s port", port)
	if err := grpcServer.Serve(lis); err != nil {
		Loge("failed to serve: %s", err)
	}
}
