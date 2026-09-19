package followrpc

import (
	"rpc-user/follow/internal/config"
	"rpc-user/follow/internal/server"
	"rpc-user/follow/internal/svc"
	"rpc-user/follow/pb"

	"google.golang.org/grpc"
)

type Config = config.Config

func Register(grpcServer *grpc.Server, c Config) {
	ctx := svc.NewServiceContext(c)
	pb.RegisterFollowServer(grpcServer, server.NewFollowServer(ctx))
}
