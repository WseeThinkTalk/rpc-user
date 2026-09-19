package userrpc

import (
	"rpc-user/user/internal/config"
	"rpc-user/user/internal/server"
	"rpc-user/user/internal/svc"
	"rpc-user/user/service"

	"google.golang.org/grpc"
)

type Config = config.Config

func Register(grpcServer *grpc.Server, c Config) {
	ctx := svc.NewServiceContext(c)
	service.RegisterUserServer(grpcServer, server.NewUserServer(ctx))
}
